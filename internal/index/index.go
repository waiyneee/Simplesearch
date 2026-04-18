package index

import (
	"fmt"
	"strings"
)

type Document struct {
	ID     int
	URL    string
	Title  string
	Body   string
	Length int
}
type Index struct {
	docTable      map[int]Document
	urlDedup      map[string]int
	invertedIndex map[string]map[int]int //for postings
	docLen        map[int]int

	// corpusStats map[int]map[int]int
	totalDocs   int
	totalDocLen int
	avgDocLen   float64
	nextDocId   int
	docFreq     map[string]int //DF per item

}

// We need a new constructor to write to nil objects inside our map
// or else program panics..
func New() *Index {
	return &Index{
		docTable:      make(map[int]Document),
		urlDedup:      make(map[string]int),
		invertedIndex: make(map[string]map[int]int),
		docLen:        make(map[int]int),

		totalDocs:   0,
		totalDocLen: 0,
		avgDocLen:   0.00,
		nextDocId:   1,
		docFreq:     make(map[string]int),
	}

}

func (idx *Index) AddDocument(url, title, body string) (docID int, isAdded bool, err error) {
	//sanity checks for nested structs and maps
	//triming url title and body
	url = strings.TrimSpace(url)
	body = strings.TrimSpace(body)

	if url == "" || body == "" {
		return 0, false, err
	}
	//deduplicacy check
	if existingId, ok := idx.urlDedup[url]; ok {
		return existingId, false, nil
	}

	tokens := Tokenize(body)
	fmt.Println(tokens)

	// if err!=nil{
	// 	return 0,false,err
	// }

	// docID=idx.nextDocId
	// idx.nextDocId++

	return

}
