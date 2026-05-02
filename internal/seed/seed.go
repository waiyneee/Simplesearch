package seed

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const apiURL = "https://en.wikipedia.org/w/api.php"

type openSearchResp []any

// ResolveWikipediaSeed returns the top Wikipedia URL for a query.
//only the top url results 
func ResolveWikipediaSeed(query string) (string, error) {
	if query == "" {
		return "", fmt.Errorf("query is empty")
	}

	u, _ := url.Parse(apiURL)
	q := u.Query()
	q.Set("action", "opensearch")
	q.Set("search", query)
	q.Set("limit", "1")
	q.Set("namespace", "0")
	q.Set("format", "json")
	q.Set("origin", "*")
	u.RawQuery = q.Encode()

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "SimpleSearchBot/0.1 (+https://github.com/waiyneee/Simplesearch)")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("wiki api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("wiki api status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var data openSearchResp
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("decode failed: %w", err)
	}

	if len(data) < 4 {
		return "", fmt.Errorf("unexpected response")
	}

	urls, ok := data[3].([]any)
	if !ok || len(urls) == 0 {
		return "", fmt.Errorf("no results")
	}

	topURL, ok := urls[0].(string)
	if !ok || topURL == "" {
		return "", fmt.Errorf("no valid url")
	}

	return topURL, nil
}