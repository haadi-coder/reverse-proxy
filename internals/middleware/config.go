package middleware

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

// ?: Почему это должно быть в пакете middleware?
type MiddlewareType string

const (
	TypeRateLimit       MiddlewareType = "rate_limit"
	TypeBasicAuth       MiddlewareType = "basic_auth"
	TypeCORS            MiddlewareType = "cors"
	TypeHeaders         MiddlewareType = "headers"
	TypeCompress        MiddlewareType = "compress"
	TypeRequestID       MiddlewareType = "request_id"
	TypeSecurityHeaders MiddlewareType = "security_headers"
)

type MiddlewareConfig struct {
	Type            MiddlewareType `yaml:"type"`
	RatelimitConfig `yaml:",inline"`
	BasicAuthConfig `yaml:",inline"`
	CORSConfig      `yaml:",inline"`
	HeadersConfig   `yaml:",inline"`
	RequestIDConfig `yaml:",inline"`
	SecurityHeaders `yaml:",inline"`
	CompressConfig  `yaml:",inline"`
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
	Remove map[string]string `yaml:"remove"`
}

type CompressConfig struct {
	MinSize int      `yaml:"min_size"`
	Level   int      `yaml:"level"`
	Types   []string `yaml:"types"`
}

type RequestIDConfig struct {
	HeaderName string `yaml:"header_name"`
}

type SecurityHeaders struct {
	ContentTypeOptions string `yaml:"content_type_options"`
	FrameOptions       string `yaml:"frame_options"`
	XSSProtection      string `yaml:"xss_protection"`
	ReferrerPolicy     string `yaml:"referrer_policy"`
	PermissionsPolicy  string `yaml:"permissions_policy"`
}

func (c *MiddlewareConfig) ApplyDefaults() {
	switch c.Type {
	case TypeBasicAuth:
		if c.Realm == "" {
			c.Realm = "Restricted"
		}

	case TypeCORS:
		if len(c.AllowedMethods) == 0 {
			c.AllowedMethods = []string{"GET", "POST"}
		}
		if c.MaxAge == 0 {
			c.MaxAge = 24 * time.Hour
		}

	case TypeCompress:
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

	case TypeRequestID:
		if c.HeaderName == "" {
			c.HeaderName = "X-Request-ID"
		}

	case TypeSecurityHeaders:
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
	types := []MiddlewareType{
		TypeBasicAuth, TypeCORS,
		TypeCompress, TypeHeaders,
		TypeRateLimit, TypeRequestID,
		TypeSecurityHeaders,
	}

	if !slices.Contains(types, c.Type) {
		return fmt.Errorf("unknown type of middleware: %s", c.Type)
	}

	switch c.Type {
	case TypeRateLimit:
		if c.Requests <= 0 {
			return fmt.Errorf("rate_limit requests must be greater then 0")
		}
		if c.Window <= 0 {
			return fmt.Errorf("rate_limit window must be greater then 0")
		}
		if c.Burst < 0 {
			return fmt.Errorf("rate_limit burst can't be negative")
		}

	case TypeBasicAuth:
		if len(c.Users) == 0 {
			return fmt.Errorf("basic_auth users is required")
		}

		for user, hash := range c.Users {
			if !isBcryptHash(hash) {
				return fmt.Errorf("basic_auth invalid bcrypt hash for user `%s`", user)
			}
		}

	case TypeCORS:
		if len(c.AllowedOrigins) == 0 {
			return fmt.Errorf("cors allowed_origins is required")
		}
		if c.MaxAge < 0 {
			return fmt.Errorf("cors max_age can't be negative")
		}

	case TypeCompress:
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
