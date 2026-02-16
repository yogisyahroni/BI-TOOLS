package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: SecurityHeaders Middleware
// ─────────────────────────────────────────────────────────────────────────────

func createTestApp(config SecurityHeadersConfig) *fiber.App {
	app := fiber.New()
	app.Use(SecurityHeaders(config))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})
	app.Get("/api/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	return app
}

func TestSecurityHeaders_CSP_SetCorrectly(t *testing.T) {
	config := DefaultSecurityHeadersConfig()
	app := createTestApp(config)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	csp := resp.Header.Get("Content-Security-Policy")
	assert.NotEmpty(t, csp)
	assert.Contains(t, csp, "default-src 'self'")
	assert.Contains(t, csp, "object-src 'none'")
	assert.Contains(t, csp, "script-src")
	assert.Contains(t, csp, "style-src")
	assert.Contains(t, csp, "img-src")
	assert.Contains(t, csp, "font-src")
}

func TestSecurityHeaders_CSP_ReportOnlyMode(t *testing.T) {
	config := DefaultSecurityHeadersConfig()
	config.ReportOnly = true
	app := createTestApp(config)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)

	// Should use Report-Only header, NOT enforcing header
	reportOnly := resp.Header.Get("Content-Security-Policy-Report-Only")
	enforcing := resp.Header.Get("Content-Security-Policy")

	assert.NotEmpty(t, reportOnly, "Report-Only CSP should be set")
	assert.Empty(t, enforcing, "Enforcing CSP should NOT be set in report-only mode")
}

func TestSecurityHeaders_XFrameOptions(t *testing.T) {
	app := createTestApp(DefaultSecurityHeadersConfig())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, "DENY", resp.Header.Get("X-Frame-Options"))
}

func TestSecurityHeaders_XContentTypeOptions(t *testing.T) {
	app := createTestApp(DefaultSecurityHeadersConfig())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"))
}

func TestSecurityHeaders_XXSSProtection(t *testing.T) {
	app := createTestApp(DefaultSecurityHeadersConfig())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	// Modern recommendation is "0" since CSP handles XSS
	assert.Equal(t, "0", resp.Header.Get("X-XSS-Protection"))
}

func TestSecurityHeaders_ReferrerPolicy(t *testing.T) {
	app := createTestApp(DefaultSecurityHeadersConfig())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, "strict-origin-when-cross-origin", resp.Header.Get("Referrer-Policy"))
}

func TestSecurityHeaders_PermissionsPolicy(t *testing.T) {
	app := createTestApp(DefaultSecurityHeadersConfig())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	pp := resp.Header.Get("Permissions-Policy")
	assert.NotEmpty(t, pp)
	assert.Contains(t, pp, "camera=()")
	assert.Contains(t, pp, "microphone=()")
}

func TestSecurityHeaders_CrossOriginPolicies(t *testing.T) {
	app := createTestApp(DefaultSecurityHeadersConfig())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)

	assert.Equal(t, "require-corp", resp.Header.Get("Cross-Origin-Embedder-Policy"))
	assert.Equal(t, "same-origin", resp.Header.Get("Cross-Origin-Opener-Policy"))
	assert.Equal(t, "same-origin", resp.Header.Get("Cross-Origin-Resource-Policy"))
}

func TestSecurityHeaders_DNSPrefetchControl(t *testing.T) {
	app := createTestApp(DefaultSecurityHeadersConfig())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, "off", resp.Header.Get("X-DNS-Prefetch-Control"))
}

func TestSecurityHeaders_DownloadOptions(t *testing.T) {
	app := createTestApp(DefaultSecurityHeadersConfig())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, "noopen", resp.Header.Get("X-Download-Options"))
}

func TestSecurityHeaders_CrossDomainPolicies(t *testing.T) {
	app := createTestApp(DefaultSecurityHeadersConfig())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, "none", resp.Header.Get("X-Permitted-Cross-Domain-Policies"))
}

func TestSecurityHeaders_APICacheControl(t *testing.T) {
	app := createTestApp(DefaultSecurityHeadersConfig())

	// Test API endpoint
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)

	cacheControl := resp.Header.Get("Cache-Control")
	assert.Contains(t, cacheControl, "no-store")
	assert.Contains(t, cacheControl, "private")
	assert.Equal(t, "no-cache", resp.Header.Get("Pragma"))
	assert.Equal(t, "0", resp.Header.Get("Expires"))
}

func TestSecurityHeaders_NonAPINoCacheHeaders(t *testing.T) {
	app := createTestApp(DefaultSecurityHeadersConfig())

	// Non-API endpoint should NOT have cache-control
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)

	cacheControl := resp.Header.Get("Cache-Control")
	assert.NotContains(t, cacheControl, "no-store")
}

func TestSecurityHeaders_ResponseBodyNotAffected(t *testing.T) {
	app := createTestApp(DefaultSecurityHeadersConfig())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "OK", string(body))
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: BuildCSPHeader
// ─────────────────────────────────────────────────────────────────────────────

func TestBuildCSPHeader_FullConfig(t *testing.T) {
	config := DefaultSecurityHeadersConfig()
	csp := config.BuildCSPHeader()

	assert.NotEmpty(t, csp)
	assert.Contains(t, csp, "default-src")
	assert.Contains(t, csp, "script-src")
	assert.Contains(t, csp, "style-src")
	assert.Contains(t, csp, "img-src")
	assert.Contains(t, csp, "font-src")
	assert.Contains(t, csp, "connect-src")
	assert.Contains(t, csp, "object-src 'none'")
	assert.Contains(t, csp, "frame-ancestors")

	// Directives separated by "; "
	parts := strings.Split(csp, "; ")
	assert.GreaterOrEqual(t, len(parts), 10)
}

func TestBuildCSPHeader_EmptyConfig(t *testing.T) {
	config := SecurityHeadersConfig{}
	csp := config.BuildCSPHeader()
	assert.Empty(t, csp)
}

func TestBuildCSPHeader_WithReportURI(t *testing.T) {
	config := DefaultSecurityHeadersConfig()
	config.ReportURI = "https://example.com/csp-report"
	csp := config.BuildCSPHeader()

	assert.Contains(t, csp, "report-uri https://example.com/csp-report")
}

func TestBuildCSPDirective_Success(t *testing.T) {
	result := buildCSPDirective("default-src", []string{"'self'", "https:"})
	assert.Equal(t, "default-src 'self' https:", result)
}

func TestBuildCSPDirective_EmptyValues(t *testing.T) {
	result := buildCSPDirective("default-src", []string{})
	assert.Empty(t, result)
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: Custom Configuration
// ─────────────────────────────────────────────────────────────────────────────

func TestSecurityHeaders_CustomConfig(t *testing.T) {
	config := SecurityHeadersConfig{
		DefaultSrc:     []string{"'self'"},
		ScriptSrc:      []string{"'self'", "https://cdn.example.com"},
		ObjectSrc:      []string{"'none'"},
		FrameAncestors: []string{"'none'"},
	}

	app := createTestApp(config)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)

	csp := resp.Header.Get("Content-Security-Policy")
	assert.Contains(t, csp, "https://cdn.example.com")
	assert.Contains(t, csp, "frame-ancestors 'none'")
}

func TestSecurityHeaders_NoCrossOriginPolicies(t *testing.T) {
	config := SecurityHeadersConfig{
		DefaultSrc: []string{"'self'"},
		// Intentionally leave cross-origin policies empty
	}

	app := createTestApp(config)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)

	assert.Empty(t, resp.Header.Get("Cross-Origin-Embedder-Policy"))
	assert.Empty(t, resp.Header.Get("Cross-Origin-Opener-Policy"))
}
