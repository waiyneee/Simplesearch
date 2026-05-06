package engine

import (
	"strings"

	"github.com/waiyneee/Simplesearch/internal/fuzzySearch/bktree"
	"github.com/waiyneee/Simplesearch/internal/fuzzySearch/levenshtein"
	"github.com/waiyneee/Simplesearch/internal/fuzzySearch/trigram"
	"github.com/waiyneee/Simplesearch/internal/fuzzySearch/types"
)

type EngineImpl struct {
	tree       *bktree.Tree
	trigrams   *trigram.Index
	words      map[string]struct{}
	maxDist    int
	limit      int
	minTrigram float64
}

func New(words []string) *EngineImpl {
	e := &EngineImpl{
		tree:       bktree.New(),
		trigrams:   trigram.New(),
		words:      map[string]struct{}{},
		maxDist:    2,
		limit:      5,
		minTrigram: 0.3,
	}
	e.Build(words)
	return e
}

func (e *EngineImpl) Build(words []string) {
	e.tree.Build(nil)
	e.trigrams.Build(nil)
	e.words = map[string]struct{}{}

	for _, w := range words {
		e.AddWord(w)
	}
}

func (e *EngineImpl) AddWord(word string) {
	w := normalize(word)
	if w == "" {
		return
	}
	if _, ok := e.words[w]; ok {
		return
	}
	e.words[w] = struct{}{}
	e.tree.Add(w)
	e.trigrams.Add(w)
}

func (e *EngineImpl) Suggest(query string, limit int) []types.Suggestion {
	q := normalize(query)
	if q == "" {
		return nil
	}

	candidates := e.trigrams.Candidates(q)
	if len(candidates) == 0 {
		return nil
	}

	if limit <= 0 {
		limit = e.limit
	}

	results := e.tree.Search(q, e.maxDist, candidates, limit)

	if len(results) == 0 {
		for _, cand := range candidates {
			d := levenshtein.Compute(q, cand)
			if d <= e.maxDist {
				results = append(results, types.Suggestion{
					Word:     cand,
					Distance: d,
					Score:    1.0 / float64(d+1),
				})
			}
		}
	}

	return results
}

func (e *EngineImpl) BestCorrection(query string) (string, bool) {
	results := e.Suggest(query, 1)
	if len(results) == 0 {
		return "", false
	}
	return results[0].Word, true
}

func normalize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.Join(strings.Fields(s), " ")
}
