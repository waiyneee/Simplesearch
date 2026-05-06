package index

import "sort"

// Terms returns all unique indexed terms.
func (idx *Index) Terms() []string {
	if idx == nil {
		return nil
	}
	terms := make([]string, 0, len(idx.invertedIndex))
	for term := range idx.invertedIndex {
		terms = append(terms, term)
	}
	sort.Strings(terms)
	return terms
}

//I need a vocabulary to work with my local+wikisuggetor as well.
