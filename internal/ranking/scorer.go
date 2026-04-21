package ranking

import (
	"github.com/waiyneee/Simplesearch/internal/index"
	"sort"
)
type SearchResult struct{
	DocID int
	Score  float64
}

type Scorer struct{
	bm25 *BM25Engine
	idx *index.Index
}

func NewScorer(idx *index.Index) *Scorer{
	if idx==nil{
		return nil
	}
    bm := NewBM25(idx)

	return &Scorer{
		bm25: bm,
		idx:  idx,
	}
}


func (s *Scorer) Score(queryTerms []string,candidateDocIDs []int,k int) []SearchResult{
   if s == nil || s.bm25 == nil || len(queryTerms) == 0 || len(candidateDocIDs) == 0 || k <= 0 {
		return nil
	}

	results:=make([]SearchResult,0,len(candidateDocIDs))

	for _,docID:=range candidateDocIDs{
		score := s.bm25.ScoreDoc(queryTerms, docID)
		if score <= 0 {
			continue
		}

		results=append(results,SearchResult{
			DocID: docID,
			Score: score,
		})

	}

	//sorting in the desc order 
	//we can use nnonymous fn here no helper  
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].DocID < results[j].DocID
		}
		return results[i].Score > results[j].Score
	})
    

	//top k results only 
	if len(results) > k {
		results = results[:k]
	}

	return results


}