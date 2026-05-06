package engine

import "github.com/waiyneee/Simplesearch/internal/fuzzySearch/types"

type Engine interface {
	Suggest(query string, limit int) []types.Suggestion
	BestCorrection(query string) (string, bool)
	AddWord(word string)
}
