package pipeline

import (
	"log"

	"github.com/waiyneee/Simplesearch/internal/index"
)

type PageToIndex struct {
	URL   string
	Title string
	Body  string
}

type IndexingOutcome struct {
	DocID int
	Added bool
	Err   error
}

// IndexPage indexes one normalized page payload.
func IndexPage(idx *index.Index, page PageToIndex) IndexingOutcome {
	docID, added, err := idx.AddDocument(page.URL, page.Title, page.Body)
	if err != nil {
		log.Printf("index error url=%s err=%v", page.URL, err)
		return IndexingOutcome{DocID: 0, Added: false, Err: err}
	}
	if !added {
		log.Printf("duplicate skipped url=%s existing_doc_id=%d", page.URL, docID)
		return IndexingOutcome{DocID: docID, Added: false, Err: nil}
	}

	log.Printf("indexed url=%s doc_id=%d", page.URL, docID)
	return IndexingOutcome{DocID: docID, Added: true, Err: nil}
}
