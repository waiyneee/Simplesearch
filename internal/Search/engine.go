package search

import (
	"github.com/waiyneee/Simplesearch/internal/index"
	"github.com/waiyneee/Simplesearch/internal/ranking"

	"fmt"
)

type Engine struct {
	idx    *index.Index
	scorer *ranking.Scorer
}

func NewEngine(idx *index.Index) *Engine {
	//constructor for engine struct
	if idx == nil {
		return nil
	}

	return &Engine{
		idx:    idx,
		scorer: ranking.NewScorer(idx),
	}

}

func (e *Engine) retrieveCandidates(queryTerms []string) []int {

	if e==nil || e.idx==nil || len(queryTerms)==0 {
		return nil
	}


	seen:=make(map[int]struct{})

	for _,term :=range queryTerms{
		postings:=e.idx.Postings(term) //add term 
		for docID:=range postings{
			seen[docID]=struct{}{}
		}
	}

	candidates:=make([]int,0,len(seen))
	for docID,_:=range seen{
		candidates=append(candidates,docID)
	}

	return candidates
}

//major search query by public/users 
func (e *Engine) Search(rawQuery string,k int) ([]ranking.SearchResult,error){
	if e==nil|| e.idx==nil || e.scorer==nil {
		return nil,fmt.Errorf("Search not initialized")

	}

	if k<=0{
		return nil,fmt.Errorf("k must be >0")

	}

	terms:=index.Tokenize(rawQuery)
	candidates:=e.retrieveCandidates(terms)
	//now tokenier helps 

	if len(terms)==0 || len(candidates)==0{
		return []ranking.SearchResult{},nil

	}
	results:=e.scorer.Score(terms,candidates,k)

	return results,nil

}

