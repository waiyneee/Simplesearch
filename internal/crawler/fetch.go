package crawler

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	userAgent   = "SimpleSearchBot/0.1 (+https://github.com/yourname/simplesearch)"
	maxBodySize = 2 << 20 // 2 MB
)

var client = &http.Client{
	Timeout: 10 * time.Second,
}

func Fetch(targetURL string) ([]byte, *http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		return nil, resp, fmt.Errorf("unexpected status code: %s", resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
	if err != nil {
		resp.Body.Close()
		return nil, resp, err
	}

	return body, resp, nil
}