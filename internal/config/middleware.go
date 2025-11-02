package config

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/haadi-coder/reverse-proxy/pkg/middleware"
)

type MiddlewareConfig struct {
	Type                  middleware.Type `yaml:"type"`
	RatelimitConfig       `yaml:",inline"`
	BasicAuthConfig       `yaml:",inline"`
	CORSConfig            `yaml:",inline"`
	HeadersConfig         `yaml:",inline"`
	RequestIDConfig       `yaml:",inline"`
	SecurityHeadersConfig `yaml:",inline"`
	CompressConfig        `yaml:",inline"`
}

type RatelimitConfig struct {
	Requests int           `yaml:"requests"`
	Window   time.Duration `yaml:"window"`
	Burst    int           `yaml:"burst"`
}

type BasicAuthConfig struct {
	Users map[string]string `yaml:"users"`
	Realm string            `yaml:"realm"`
}

type CORSConfig struct {
	AllowedOrigins   []string      `yaml:"allowed_origins"`
	AllowedMethods   []string      `yaml:"allowed_methods"`
	AllowedHeaders   []string      `yaml:"allowed_headers"`
	ExposedHeaders   []string      `yaml:"exposed_headers"`
	AllowCredentials bool          `yaml:"allow_credentials"`
	MaxAge           time.Duration `yaml:"max_age"`
}

type HeadersConfig struct {
	Request  *HeaderRules `yaml:"request"`
	Response *HeaderRules `yaml:"response"`
}
type HeaderRules struct {
	Add    map[string]string `yaml:"add"`
	Set    map[string]string `yaml:"set"`
	Remove []string          `yaml:"remove"`
}

type CompressConfig struct {
	MinSize int      `yaml:"min_size"`
	Level   int      `yaml:"level"`
	Types   []string `yaml:"types"`
}

type RequestIDConfig struct {
	HeaderName string `yaml:"header_name"`
}

type SecurityHeadersConfig struct {
	ContentTypeOptions string `yaml:"content_type_options"`
	FrameOptions       string `yaml:"frame_options"`
	XSSProtection      string `yaml:"xss_protection"`
	ReferrerPolicy     string `yaml:"referrer_policy"`
	PermissionsPolicy  string `yaml:"permissions_policy"`
}

func (c *MiddlewareConfig) ApplyDefaults() {
	switch c.Type {
	case middleware.TypeBasicAuth:
		if c.Realm == "" {
			c.Realm = "Restricted"
		}

	case middleware.TypeCORS:
		if len(c.AllowedMethods) == 0 {
			c.AllowedMethods = []string{"GET", "POST"}
		}
		if c.MaxAge == 0 {
			c.MaxAge = 24 * time.Hour
		}

	case middleware.TypeCompress:
		if c.MinSize == 0 {
			c.MinSize = 1024
		}
		if c.Level == 0 {
			c.Level = 5
		}
		if len(c.Types) == 0 {
			c.Types = []string{
				"text/html",
				"text/css",
				"text/javascript",
				"application/javascript",
				"application/json",
				"text/xml",
				"application/xml",
				"application/rss+xml",
				"image/svg+xml",
			}
		}

	case middleware.TypeRequestID:
		if c.HeaderName == "" {
			c.HeaderName = "X-Request-ID"
		}

	case middleware.TypeSecurityHeaders:
		if c.ContentTypeOptions == "" {
			c.ContentTypeOptions = "nosniff"
		}
		if c.FrameOptions == "" {
			c.FrameOptions = "DENY"
		}
		if c.XSSProtection == "" {
			c.XSSProtection = "1; mode=block"
		}
		if c.ReferrerPolicy == "" {
			c.ReferrerPolicy = "strict-origin-when-cross-origin"
		}
	}
}

func (c *MiddlewareConfig) Validate() error {
	types := []middleware.Type{
		middleware.TypeBasicAuth, middleware.TypeCORS,
		middleware.TypeCompress, middleware.TypeHeaders,
		middleware.TypeRateLimit, middleware.TypeRequestID,
		middleware.TypeSecurityHeaders,
	}

	if !slices.Contains(types, c.Type) {
		return fmt.Errorf("unknown type of middleware: %s", c.Type)
	}

	switch c.Type {
	case middleware.TypeRateLimit:
		if c.Requests <= 0 {
			return fmt.Errorf("rate_limit requests must be greater then 0")
		}
		if c.Window <= 0 {
			return fmt.Errorf("rate_limit window must be greater then 0")
		}
		if c.Burst < 0 {
			return fmt.Errorf("rate_limit burst can't be negative")
		}

	case middleware.TypeBasicAuth:
		if len(c.Users) == 0 {
			return fmt.Errorf("basic_auth users is required")
		}

		for user, hash := range c.Users {
			if !isBcryptHash(hash) {
				return fmt.Errorf("basic_auth invalid bcrypt hash for user `%s`", user)
			}
		}

	case middleware.TypeCORS:
		if len(c.AllowedOrigins) == 0 {
			return fmt.Errorf("cors allowed_origins is required")
		}
		if c.MaxAge < 0 {
			return fmt.Errorf("cors max_age can't be negative")
		}

	case middleware.TypeCompress:
		if c.MinSize < 0 {
			return fmt.Errorf("compress min_size can't be negative")
		}
		if c.Level < 1 || c.Level > 9 {
			return fmt.Errorf("compress level must be between 1 and 9")
		}
	}

	return nil
}

func isBcryptHash(hash string) bool {
	return strings.HasPrefix(hash, "$2a") ||
		strings.HasPrefix(hash, "$2b") ||
		strings.HasPrefix(hash, "$2y")
}

func (c *MiddlewareConfig) Build() (*middleware.Middleware, error) {
	switch c.Type {
	case middleware.TypeRateLimit:
		return middleware.RateLimit(&middleware.RatelimitConfig{
			Requests: c.Requests,
			Window:   c.Window,
			Burst:    c.Burst,
		}), nil

	case middleware.TypeBasicAuth:
		return middleware.BasicAuth(&middleware.BasicAuthConfig{
			Users: c.Users,
			Realm: c.Realm,
		}), nil

	case middleware.TypeCORS:
		return middleware.CORS(&middleware.CORSConfig{
			AllowedOrigins:   c.AllowedOrigins,
			AllowedMethods:   c.AllowedMethods,
			AllowedHeaders:   c.AllowedHeaders,
			ExposedHeaders:   c.ExposedHeaders,
			AllowCredentials: c.AllowCredentials,
			MaxAge:           c.MaxAge,
		}), nil

	case middleware.TypeHeaders:
		return middleware.Headers(&middleware.HeadersConfig{
			Request: &middleware.HeaderRules{
				Add:    c.Request.Add,
				Set:    c.Request.Set,
				Remove: c.Request.Remove,
			},
			Response: &middleware.HeaderRules{
				Add:    c.Response.Add,
				Set:    c.Response.Set,
				Remove: c.Response.Remove,
			},
		}), nil

	case middleware.TypeRequestID:
		return middleware.RequestID(&middleware.RequestIDConfig{
			HeaderName: c.HeaderName,
		}), nil

	case middleware.TypeSecurityHeaders:
		return middleware.SecurityHeaders(&middleware.SecurityHeadersConfig{
			ContentTypeOptions: c.ContentTypeOptions,
			FrameOptions:       c.FrameOptions,
			XSSProtection:      c.XSSProtection,
			ReferrerPolicy:     c.ReferrerPolicy,
			PermissionsPolicy:  c.PermissionsPolicy,
		}), nil

	case middleware.TypeCompress:
		return middleware.Compress(&middleware.CompressConfig{
			MinSize: c.MinSize,
			Level:   c.Level,
			Types:   c.Types,
		}), nil

	default:
		return nil, fmt.Errorf("unknow middleware type: %s", c.Type)
	}
}
