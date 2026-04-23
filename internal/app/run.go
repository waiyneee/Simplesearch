package app

import (
	"strings"

	"github.com/waiyneee/Simplesearch/internal/ranking"
)

// SearchRequest is the app-level input contract.
type SearchRequest struct {
	Query string
	TopK  int
}

// SearchResponse is the app-level output contract.
type SearchResponse struct {
	Results []ranking.SearchResult
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

	results, err := a.engine.Search(req.Query, req.TopK)
	if err != nil {
		return SearchResponse{}, err
	}

	return SearchResponse{
		Results: results,
	}, nil
}
