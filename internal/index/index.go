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

// DocCount returns total number of indexed documents.
func (idx *Index) DocCount() int {
	return idx.totalDocs
}

// AvgDocLength returns average token length across indexed documents.
func (idx *Index) AvgDocLength() float64 {
	return idx.avgDocLen
}

// DocLength returns token length for a document ID.
// Returns 0 if docID is not found.
func (idx *Index) DocLength(docID int) int {
	return idx.docLen[docID]
}

// DocumentFrequency returns in how many docs this term appears.
func (idx *Index) DocumentFrequency(term string) int {
	term = strings.TrimSpace(term)
	if term == "" {
		return 0
	}
	return idx.docFreq[term]
}

// TermFrequency returns frequency of term inside a specific document.
// Returns 0 if term/doc not found.
func (idx *Index) TermFrequency(term string, docID int) int {
	term = strings.TrimSpace(term)
	if term == "" || docID <= 0 {
		return 0
	}

	postings, ok := idx.invertedIndex[term]
	if !ok {
		return 0
	}
	return postings[docID]
}
func (idx *Index) Postings(term string) map[int]int {
	term = strings.TrimSpace(term)
	if term == "" {
		return map[int]int{}
		//empty
	}

	postings, ok := idx.invertedIndex[term]
	if !ok {
		return map[int]int{}
	}

	return postings

}
