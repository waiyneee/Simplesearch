package search

import (
	"github.com/waiyneee/Simplesearch/internal/index"

	"github.com/waiyneee/Simplesearch/internal/ranking"
)

type Engine struct {
	idx    *index.Index
	scorer *ranking.Scorer
}

func NewEngine(idx *index.Index) *Engine {
	if idx == nil {
		return nil
	}

	sc := ranking.NewScorer(idx)
	if sc == nil {
		return nil
	}

	return &Engine{
		idx:    idx,
		scorer: sc,
	}
}

// Search executes: parse/validate -> candidate match -> rank.
func (e *Engine) Search(rawQuery string, k int) ([]ranking.SearchResult, error) {
	if e == nil || e.idx == nil || e.scorer == nil {
		return nil, ErrEngineNotInit
	}
	if k <= 0 {
		return nil, ErrInvalidTopK
	}

	parsed, err := ParseAndValidateQuery(rawQuery)
	if err != nil {
		// For user-facing search UX, invalid/empty query is not a server fault.
		// Return empty results + explicit error for caller decision.
		return []ranking.SearchResult{}, err
	}

	candidates := e.retrieveCandidates(parsed.Terms)
	if len(candidates) == 0 {
		return []ranking.SearchResult{}, nil
	}

	results := e.scorer.Score(parsed.Terms, candidates, k)
	if len(results) == 0 {
		return []ranking.SearchResult{}, nil
	}

	return results, nil
}
