package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type ScreenshotService struct {
	BaseURL string
}

func NewScreenshotService() *ScreenshotService {
	baseURL := os.Getenv("FRONTEND_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}
	return &ScreenshotService{
		BaseURL: baseURL,
	}
}

// CaptureDashboard takes a screenshot of a specific dashboard
// token: JWT token for authentication
func (s *ScreenshotService) CaptureDashboard(dashboardID string, token string, width, height int64) ([]byte, error) {
	// Create context
	// If running effectively in a container without a display, use Headless mode
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.WindowSize(int(width), int(height)),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// Create context with timeout
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Set timeout for the entire operation
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var buf []byte
	url := fmt.Sprintf("%s/dashboards/%s?mode=print", s.BaseURL, dashboardID)

	// Auth cookie setup
	// We assume the frontend looks for a specific cookie or local storage
	// For NextAuth, it's usually "next-auth.session-token"
	authCookie := &network.CookieParam{
		Name:   "next-auth.session-token",
		Value:  token,       // We might need to handle token/session consistency
		Domain: "localhost", // TODO: Make dynamic
		Path:   "/",
	}

	// Local storage injection script if cookie doesn't work (for JWT stored in LS)
	// setTokenScript := fmt.Sprintf(`localStorage.setItem("token", "%s");`, token)

	log.Printf("Capturing screenshot for dashboard %s at %s", dashboardID, url)

	err := chromedp.Run(ctx,
		// Init network to set cookies
		network.Enable(),
		network.SetCookie(authCookie.Name, authCookie.Value).
			WithDomain(authCookie.Domain).
			WithPath(authCookie.Path),

		chromedp.Navigate(url),
		// Wait for data to load - specific selector for dashboard grid or a timeout
		// chromedp.WaitVisible(".dashboard-grid", chromedp.ByQuery), // Hypothetical selector
		chromedp.Sleep(5*time.Second), // Simple fallback wait
		chromedp.CaptureScreenshot(&buf),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to capture screenshot: %w", err)
	}

	return buf, nil
}
