package index

// db helpers
func (idx *Index) DocTable() map[int]Document {
	return idx.docTable
}

func (idx *Index) InvertedIndex() map[string]map[int]int {
	return idx.invertedIndex
}

func (idx *Index) TotalDocs() int {
	return idx.totalDocs
}

func (idx *Index) TotalDocLen() int {
	return idx.totalDocLen
}

func (idx *Index) NextDocID() int {
	return idx.nextDocId
}

// Insert a document directly (no tokenization)
func (idx *Index) AddDocumentFromDB(doc Document) {
	idx.docTable[doc.ID] = doc
	idx.urlDedup[doc.URL] = doc.ID
	idx.docLen[doc.ID] = doc.Length
	if doc.ID >= idx.nextDocId {
		idx.nextDocId = doc.ID + 1
	}
}

// Insert a posting directly (no recompute)
func (idx *Index) AddPostingFromDB(term string, docID int, tf int) {
	if idx.invertedIndex[term] == nil {
		idx.invertedIndex[term] = make(map[int]int)
	}
	idx.invertedIndex[term][docID] = tf
	idx.docFreq[term]++
}

// Set stats after load
func (idx *Index) SetStatsFromDB(totalDocs, totalDocLen, nextDocID int) {
	idx.totalDocs = totalDocs
	idx.totalDocLen = totalDocLen
	if totalDocs > 0 {
		idx.avgDocLen = float64(totalDocLen) / float64(totalDocs)
	}
	idx.nextDocId = nextDocID
}
