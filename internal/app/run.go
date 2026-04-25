package app

import (
	"strings"
)

// SearchRequest is the app-level input contract.
type SearchRequest struct {
	Query string
	TopK  int
}

// SearchResultView is user-facing model needed enriched output.
type SearchResultView struct {
	DocID   int
	Score   float64
	Title   string
	URL     string
	Snippet string
}

// SearchResponse is the app-level output contract.something to show to users
type SearchResponse struct {
	Results []SearchResultView
}

// Run executes one search request end-to-end via search.Engine.
func (a *App) Run(req SearchRequest) (SearchResponse, error) {
	if a == nil || a.idx == nil || a.engine == nil {
		return SearchResponse{}, ErrEngineInit
	}

	req.Query = strings.TrimSpace(req.Query)
	if req.Query == "" {
		return SearchResponse{}, ErrEmptyQuery
	}
	if req.TopK <= 0 {
		return SearchResponse{}, ErrInvalidTopK
	}

	raw, err := a.engine.Search(req.Query, req.TopK)
	if err != nil {
		return SearchResponse{}, err
	}

	enriched := make([]SearchResultView, 0, len(raw))
	for _, r := range raw {
		doc, ok := a.idx.GetDocument(r.DocID)
		if !ok {
			continue
		}

		enriched = append(enriched, SearchResultView{
			DocID:   r.DocID,
			Score:   r.Score,
			Title:   doc.Title,
			URL:     doc.URL,
			Snippet: buildSnippet(doc.Body, 500),
		})
	}

	return SearchResponse{Results: enriched}, nil
}

func buildSnippet(body string, maxChars int) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}
	body = strings.Join(strings.Fields(body), " ")
	if maxChars <= 0 || len(body) <= maxChars {
		return body
	}
	return body[:maxChars] + "..."
}
