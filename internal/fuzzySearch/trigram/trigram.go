package trigram

import "strings"

type TrigramIndex interface {
	Build(words []string)
	Add(word string)
	Candidates(query string) []string
}

type Index struct {
	index map[string]map[string]struct{} //inverted index type
}

func New() *Index { //constructor
	return &Index{index: map[string]map[string]struct{}{}}
}

func (t *Index) Build(words []string) {
	t.index = map[string]map[string]struct{}{}
	for _, w := range words {
		t.Add(w)
	}
}

func (t *Index) Add(word string) {
	word = strings.ToLower(strings.TrimSpace(word))
	if word == "" {
		return
	}
	for _, tri := range trigrams(word) {
		if t.index[tri] == nil {
			t.index[tri] = map[string]struct{}{}
		}
		t.index[tri][word] = struct{}{}
	}
}

func (t *Index) Candidates(query string) []string {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return nil
	}

	seen := map[string]struct{}{}
	for _, tri := range trigrams(query) {
		for w := range t.index[tri] {
			seen[w] = struct{}{}
		}
	}

	out := make([]string, 0, len(seen))
	for w := range seen {
		out = append(out, w)
	}
	return out
}

func trigrams(word string) []string {
	if len(word) < 3 {
		return []string{word}
	}
	// pad with ^ and $
	padded := "^" + word + "$"
	out := make([]string, 0, len(padded)-2)
	for i := 0; i <= len(padded)-3; i++ {
		out = append(out, padded[i:i+3])
	}
	return out
}
