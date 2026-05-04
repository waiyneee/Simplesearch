package TrigramIndex

type TrigramIndex interface {
	Build(words []string) //building the index
	Add(word string)

	Candidates(query string) []string
	// candidates returns words that share trigrams with query.

}
