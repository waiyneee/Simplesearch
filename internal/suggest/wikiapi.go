package suggest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

type WikiHTTP struct {
	client    *http.Client
	baseURL   string
	userAgent string
}

//search api usage for title correction needed...

func NewWikiSuggestor() *WikiHTTP {
	return &WikiHTTP{
		client: &http.Client{
			Timeout: 4 * time.Second,
		},
		baseURL:   "https://en.wikipedia.org/w/rest.php/v1/search/title",
		userAgent: "SimpleSearchBot/0.1 (+https://github.com/waiyneee/Simplesearch)",
	}
}

func (w *WikiHTTP) TopTitle(query string) (string, error) {
	return w.TopTitleWithContext(context.Background(), query)
}

func (w *WikiHTTP) TopTitleWithContext(ctx context.Context, query string) (string, error) {
	if w == nil || query == "" {
		return "", nil
	}

	q := url.Values{}
	q.Set("q", query)
	q.Set("limit", "1")

	req, err := http.NewRequestWithContext(ctx, "GET", w.baseURL+"?"+q.Encode(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", w.userAgent)

	resp, err := w.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var payload struct {
		Pages []struct {
			Title string `json:"title"`
		} `json:"pages"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}

	if len(payload.Pages) == 0 {
		return "", nil
	}
	return payload.Pages[0].Title, nil
}
