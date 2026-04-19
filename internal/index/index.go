package index

import (
	"fmt"
	"strings"
)

// New initializes all maps/counters to avoid nil-map panics.
func New() *Index {
	return &Index{
		docTable:      make(map[int]Document),
		urlDedup:      make(map[string]int),
		invertedIndex: make(map[string]map[int]int),
		docLen:        make(map[int]int),
		totalDocs:     0,
		totalDocLen:   0,
		avgDocLen:     0,
		nextDocId:     1,
		docFreq:       make(map[string]int),
	}
}

func (idx *Index) AddDocument(url, title, body string) (docID int, isAdded bool, err error) {
	url = strings.TrimSpace(url)
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)

	if url == "" {
		return 0, false, fmt.Errorf("add document: url is empty")
	}
	if body == "" {
		return 0, false, fmt.Errorf("add document: body is empty")
	}

	// URL dedup
	if existingID, ok := idx.urlDedup[url]; ok {
		return existingID, false, nil
	}

	tokens := Tokenize(body)
	if len(tokens) == 0 {
		return 0, false, fmt.Errorf("add document: no indexable tokens after normalization")
	}

	// Assign new doc ID
	docID = idx.nextDocId
	idx.nextDocId++

	// Build TF map (term -> frequency in this document)
	tfMap := buildTermFreq(tokens)

	// Store document
	doc := Document{
		ID:     docID,
		URL:    url,
		Title:  title,
		Body:   body,
		Length: len(tokens),
	}
	idx.docTable[docID] = doc
	idx.urlDedup[url] = docID
	idx.docLen[docID] = len(tokens)

	// Update inverted index and DF
	idx.applyPostings(docID, tfMap)

	// Update corpus stats
	idx.totalDocs++
	idx.totalDocLen += len(tokens)
	idx.avgDocLen = float64(idx.totalDocLen) / float64(idx.totalDocs)

	return docID, true, nil
}
