package search

import "sort"


// retrieveCandidates performs OR matching across terms and returns unique docIDs.
func (e *Engine) retrieveCandidates(queryTerms []string) []int {
	if e == nil || e.idx == nil || len(queryTerms) == 0 {
		return []int{}
	}

	seenDocs := make(map[int]struct{}, 256)

	for _, term := range queryTerms {
		if term == "" {
			continue
		}
		postings := e.idx.Postings(term) // expected read-only use
		for docID := range postings {
			if docID <= 0 {
				continue
			}
			seenDocs[docID] = struct{}{}
		}
	}

	if len(seenDocs) == 0 {
		return []int{}
	}

	candidates := make([]int, 0, len(seenDocs))
	for docID := range seenDocs {
		candidates = append(candidates, docID)
	}

	sort.Ints(candidates)
	return candidates
}