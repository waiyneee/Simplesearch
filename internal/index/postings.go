package index

// type Posting struct{
// 	DocID int
// 	TF int //tf==term frequency ::frequency of a term in single document 
// }

func buildTermFreq(tokens []string) map[string]int{

  tfMap := make(map[string]int, len(tokens))
	for _, term := range tokens {
		tfMap[term]++
	}

	//after building tf assigning it.

	return tfMap

}

func (idx *Index) addPosting(term string,docID int,tf int){
	if idx.invertedIndex[term]==nil{
		idx.invertedIndex[term]=make(map[int]int)

	}
     idx.invertedIndex[term][docID]=tf

}


func (idx *Index) applyPostings(docID int,tfMap map[string]int){

	for term,termfreq:=range tfMap{
		idx.addPosting(term,docID,termfreq)
		idx.docFreq[term]++

	}
	
}

