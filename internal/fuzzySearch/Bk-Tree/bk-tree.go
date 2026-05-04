package bktree

import "github.com/waiyneee/Simplesearch/internal/fuzzySearch/engine"

type BKtree interface {
	Build(words []string) //building atree from a word list
	//for n edits computation

	Add(word string) //adds or inserts  single word

	Search(query string, maxDistance int, candidates []string, limit int) []engine.Suggestion
	// Search returns matches within maxDist, optionally filtered by candidates.

}
