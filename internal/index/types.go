package index


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
	invertedIndex map[string]map[int]int // term -> (docID -> tf)
	docLen        map[int]int

	totalDocs   int
	totalDocLen int
	avgDocLen   float64
	nextDocId   int
	docFreq     map[string]int // term -> df
}