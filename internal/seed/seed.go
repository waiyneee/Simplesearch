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

// ResolveWikipediaSeed returns the best Wikipedia URL for a query,
// using title-boost scoring (exact > prefix > substring).
func ResolveWikipediaSeed(query string) (string, error) {
	if query == "" {
		return "", fmt.Errorf("query is empty")
	}

	u, _ := url.Parse(apiURL)
	q := u.Query()
	q.Set("action", "opensearch")
	q.Set("search", query)
	q.Set("limit", "10")
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

	titles, ok := data[1].([]any)
	if !ok || len(titles) == 0 {
		return "", fmt.Errorf("no titles")
	}

	urls, ok := data[3].([]any)
	if !ok || len(urls) == 0 {
		return "", fmt.Errorf("no results")
	}

	// Choose best URL via title-boost scoring.
	queryNorm := normalize(query)
	bestScore := -1
	bestRank := 1<<30
	bestURL := ""

	for i := 0; i < len(urls); i++ {
		if i >= len(titles) {
			break
		}
		title, ok1 := titles[i].(string)
		link, ok2 := urls[i].(string)
		if !ok1 || !ok2 || link == "" {
			continue
		}

		score := scoreTitle(queryNorm, normalize(title))

		// Higher score wins; tie-breaker is original rank (lower index).
		if score > bestScore || (score == bestScore && i < bestRank) {
			bestScore = score
			bestRank = i
			bestURL = link
		}
	}

	if bestURL == "" {
		return "", fmt.Errorf("no valid url")
	}

	return bestURL, nil
}

func normalize(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, "_", " ")
	return s
}

func scoreTitle(query, title string) int {
	switch {
	case title == query:
		return 100 // exact match boost
	case strings.HasPrefix(title, query):
		return 75 // prefix match boost
	case strings.Contains(title, query):
		return 50 // substring match boost
	default:
		return 0
	}
}