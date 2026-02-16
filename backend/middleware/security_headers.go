package middleware

import (
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// SecurityHeadersConfig configures Content-Security-Policy and advanced security headers.
// TASK-174: CSP Implementation
// TASK-175: Advanced Security Headers Middleware
type SecurityHeadersConfig struct {
	// CSP Directives
	DefaultSrc     []string // default-src
	ScriptSrc      []string // script-src
	StyleSrc       []string // style-src
	ImgSrc         []string // img-src
	FontSrc        []string // font-src
	ConnectSrc     []string // connect-src
	FrameSrc       []string // frame-src
	ObjectSrc      []string // object-src
	MediaSrc       []string // media-src
	WorkerSrc      []string // worker-src
	ChildSrc       []string // child-src
	BaseURI        []string // base-uri
	FormAction     []string // form-action
	FrameAncestors []string // frame-ancestors

	// Report URI for CSP violations
	ReportURI string

	// Extra headers
	PermissionsPolicy   string // Permissions-Policy header
	CrossOriginEmbedder string // Cross-Origin-Embedder-Policy
	CrossOriginOpener   string // Cross-Origin-Opener-Policy
	CrossOriginResource string // Cross-Origin-Resource-Policy

	// Toggle CSP report-only mode (for testing)
	ReportOnly bool
}

// DefaultSecurityHeadersConfig returns a production-ready CSP configuration.
// This is intentionally strict — you loosen as needed per endpoint.
func DefaultSecurityHeadersConfig() SecurityHeadersConfig {
	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "http://localhost:3000"
	}

	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}

	return SecurityHeadersConfig{
		DefaultSrc:     []string{"'self'"},
		ScriptSrc:      []string{"'self'", "'unsafe-inline'", "'unsafe-eval'"}, // unsafe-eval needed for Next.js dev
		StyleSrc:       []string{"'self'", "'unsafe-inline'", "https://fonts.googleapis.com"},
		ImgSrc:         []string{"'self'", "data:", "blob:", "https:"},
		FontSrc:        []string{"'self'", "https://fonts.gstatic.com", "data:"},
		ConnectSrc:     []string{"'self'", apiURL, appURL, "wss:", "ws:"},
		FrameSrc:       []string{"'self'"},
		ObjectSrc:      []string{"'none'"},
		MediaSrc:       []string{"'self'"},
		WorkerSrc:      []string{"'self'", "blob:"},
		ChildSrc:       []string{"'self'", "blob:"},
		BaseURI:        []string{"'self'"},
		FormAction:     []string{"'self'"},
		FrameAncestors: []string{"'self'"},

		PermissionsPolicy:   "camera=(), microphone=(), geolocation=(self), payment=(), usb=(), magnetometer=()",
		CrossOriginEmbedder: "require-corp",
		CrossOriginOpener:   "same-origin",
		CrossOriginResource: "same-origin",

		ReportOnly: false,
	}
}

// buildCSPDirective constructs a single CSP directive string.
func buildCSPDirective(name string, values []string) string {
	if len(values) == 0 {
		return ""
	}
	return name + " " + strings.Join(values, " ")
}

// BuildCSPHeader constructs the full Content-Security-Policy header value.
func (c SecurityHeadersConfig) BuildCSPHeader() string {
	directives := []string{}

	pairs := []struct {
		name   string
		values []string
	}{
		{"default-src", c.DefaultSrc},
		{"script-src", c.ScriptSrc},
		{"style-src", c.StyleSrc},
		{"img-src", c.ImgSrc},
		{"font-src", c.FontSrc},
		{"connect-src", c.ConnectSrc},
		{"frame-src", c.FrameSrc},
		{"object-src", c.ObjectSrc},
		{"media-src", c.MediaSrc},
		{"worker-src", c.WorkerSrc},
		{"child-src", c.ChildSrc},
		{"base-uri", c.BaseURI},
		{"form-action", c.FormAction},
		{"frame-ancestors", c.FrameAncestors},
	}

	for _, p := range pairs {
		d := buildCSPDirective(p.name, p.values)
		if d != "" {
			directives = append(directives, d)
		}
	}

	if c.ReportURI != "" {
		directives = append(directives, fmt.Sprintf("report-uri %s", c.ReportURI))
	}

	return strings.Join(directives, "; ")
}

// SecurityHeaders returns a Fiber middleware that sets CSP and all advanced security headers.
// This complements the existing SSLRedirect middleware which handles HSTS and basic XSS/Content-Type headers.
//
// Headers set:
//   - Content-Security-Policy (or Content-Security-Policy-Report-Only)
//   - Permissions-Policy
//   - Cross-Origin-Embedder-Policy
//   - Cross-Origin-Opener-Policy
//   - Cross-Origin-Resource-Policy
//   - X-Content-Type-Options (nosniff)
//   - X-Frame-Options (DENY)
//   - X-XSS-Protection (0 — deprecated, CSP handles this now)
//   - X-DNS-Prefetch-Control (off)
//   - X-Download-Options (noopen)
//   - X-Permitted-Cross-Domain-Policies (none)
//   - Referrer-Policy (strict-origin-when-cross-origin)
//   - Cache-Control (for API responses)
func SecurityHeaders(config SecurityHeadersConfig) fiber.Handler {
	cspHeader := config.BuildCSPHeader()

	return func(c *fiber.Ctx) error {
		// ── Content Security Policy ──────────────────────────────────────
		if cspHeader != "" {
			if config.ReportOnly {
				c.Set("Content-Security-Policy-Report-Only", cspHeader)
			} else {
				c.Set("Content-Security-Policy", cspHeader)
			}
		}

		// ── Standard Security Headers ────────────────────────────────────
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "0") // Modern CSP supersedes this; "0" is recommended
		c.Set("X-DNS-Prefetch-Control", "off")
		c.Set("X-Download-Options", "noopen")
		c.Set("X-Permitted-Cross-Domain-Policies", "none")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// ── Cross-Origin Policies ────────────────────────────────────────
		if config.CrossOriginEmbedder != "" {
			c.Set("Cross-Origin-Embedder-Policy", config.CrossOriginEmbedder)
		}
		if config.CrossOriginOpener != "" {
			c.Set("Cross-Origin-Opener-Policy", config.CrossOriginOpener)
		}
		if config.CrossOriginResource != "" {
			c.Set("Cross-Origin-Resource-Policy", config.CrossOriginResource)
		}

		// ── Permissions Policy ───────────────────────────────────────────
		if config.PermissionsPolicy != "" {
			c.Set("Permissions-Policy", config.PermissionsPolicy)
		}

		// ── API Cache Control ────────────────────────────────────────────
		// For API endpoints, prevent caching of sensitive data
		path := c.Path()
		if strings.HasPrefix(path, "/api/") {
			c.Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
			c.Set("Pragma", "no-cache")
			c.Set("Expires", "0")
		}

		return c.Next()
	}
}
