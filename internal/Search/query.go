package search

import (
	"github.com/waiyneee/Simplesearch/internal/index"

	"strings"
)

func NormalizeQuery(rawText string) []string {
	rawText = strings.TrimSpace(rawText)
	if rawText == "" {
		return []string{}
	}

	tokens := index.Tokenize(rawText)
	if len(tokens) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(tokens))
	terms := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		terms = append(terms, t)
	}
	return terms
}