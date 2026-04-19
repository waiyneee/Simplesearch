package pipeline

import (
	"log"

	"github.com/waiyneee/Simplesearch/internal/crawler"
	"github.com/waiyneee/Simplesearch/internal/index"
)

type IndexingOutcome struct {
	DocID int
	Added bool
	Err   error
}

// IndexPage connects crawler output to index ingestion.
func IndexPage(idx *index.Index, page crawler.PageResult) IndexingOutcome {
	docID, added, err := idx.AddDocument(page.URL, page.Title, page.BodyText)
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