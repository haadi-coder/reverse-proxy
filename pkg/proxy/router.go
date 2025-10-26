package proxy

import (
	"strings"
	"sync"
)

type Router struct {
	exact     map[string]*Route
	wildcards map[string]*Route
	mu        sync.RWMutex
}

func (r *Router) add(host string, route *Route) {
	r.mu.Lock()
	defer r.mu.Unlock()

	prepared := prepareHost(host)

	if isWildcard(prepared) {
		r.wildcards[prepared] = route
	} else {
		r.exact[prepared] = route
	}
}

func (r *Router) remove(host string) {
	prepared := prepareHost(host)

	r.mu.Lock()
	defer r.mu.Unlock()

	if isWildcard(prepared) {
		delete(r.wildcards, prepared)
	} else {
		delete(r.exact, prepared)
	}
}

func (r *Router) lookup(host string) (*Route, bool) {
	prepared := prepareHost(host)

	r.mu.RLock()
	defer r.mu.RUnlock()

	if route, ok := r.exact[prepared]; ok {
		return route, true
	}

	for pattern, route := range r.wildcards {
		if matchWildcard(prepared, pattern) {
			return route, true
		}
	}

	return nil, false
}

func prepareHost(host string) string {
	host = strings.ToLower(host)

	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	return host
}

func matchWildcard(host, pattern string) bool {
	if !isWildcard(pattern) {
		return false
	}

	suffix := pattern[1:]
	return strings.HasSuffix(host, suffix)
}

func isWildcard(host string) bool {
	return strings.HasPrefix(host, "*")
}
