package crawler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultUserAgent = "SimpleSearchBot/0.1 (+https://github.com/waiyneee/Simplesearch)"
)

var httpClient = &http.Client{
	Timeout: 12 * time.Second,
}

// Fetch downloads a URL and returns response body + status code.
// maxBytes controls max response body bytes read from this URL.
// It only accepts HTML/XHTML content types.
func Fetch(ctx context.Context, targetURL, userAgent string, maxBytes int64) ([]byte, int, error) {
	if userAgent == "" {
		userAgent = defaultUserAgent
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, resp.StatusCode, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if contentType == "" {
		return nil, resp.StatusCode, fmt.Errorf("missing content-type")
	}
	if !strings.Contains(contentType, "text/html") && !strings.Contains(contentType, "application/xhtml+xml") {
		return nil, resp.StatusCode, fmt.Errorf("unsupported content-type: %s", contentType)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes))
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return body, resp.StatusCode, nil
}