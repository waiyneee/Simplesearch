package ranking

import (
	"sort"
	"strings"

	"github.com/waiyneee/Simplesearch/internal/index"
)

type SearchResult struct {
	DocID int
	Score float64
}

type Scorer struct {
	bm25 *BM25Engine
	idx  *index.Index
}

func NewScorer(idx *index.Index) *Scorer {
	if idx == nil {
		return nil
	}
	bm := NewBM25(idx)

	return &Scorer{
		bm25: bm,
		idx:  idx,
	}
}

func (s *Scorer) Score(queryTerms []string, candidateDocIDs []int, k int) []SearchResult {
	if s == nil || s.bm25 == nil || len(queryTerms) == 0 || len(candidateDocIDs) == 0 || k <= 0 {
		return nil
	}

	results := make([]SearchResult, 0, len(candidateDocIDs))

	queryNorm := normalizeTitle(strings.Join(queryTerms, " "))

	for _, docID := range candidateDocIDs {
		score := s.bm25.ScoreDoc(queryTerms, docID)
		if score <= 0 {
			continue
		}

		if doc, ok := s.idx.GetDocument(docID); ok {
			boost := titleBoost(queryNorm, normalizeTitle(doc.Title))
			score *= boost
		}

		results = append(results, SearchResult{
			DocID: docID,
			Score: score,
		})
	}

	// sorting in desc order
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].DocID < results[j].DocID
		}
		return results[i].Score > results[j].Score
	})

	// top k results only
	if len(results) > k {
		results = results[:k]
	}

	return results
}

func normalizeTitle(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, "_", " ")
	return s
}

// exact > prefix > substring
func titleBoost(query, title string) float64 {
	switch {
	case query == "" || title == "":
		return 1.0
	case title == query:
		return 1.75
	case strings.HasPrefix(title, query):
		return 1.35
	case strings.Contains(title, query):
		return 1.15
	default:
		return 1.0
	}
}
