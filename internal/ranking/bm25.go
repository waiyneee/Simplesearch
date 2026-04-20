package ranking


import (
	"math"
    "github.com/waiyneee/Simplesearch/internal/index"
)
 
type Bm25Defaults interface{
	defaults()
}
type BM25Engine struct{
	idx *index.Index 
	k1 float64
	b float64
}

func (d *BM25Engine) defaults() {
	d.k1 = 1.2
	d.b = 0.75
}

func idf(df int,N int) float64{

	if df <= 0 || N <= 0 {
		return 0
	}
	return math.Log(1.0+(float64(N-df)+0.5)/float64(df)+0.5)

}

func tfNormalized(tf,docLen int,avgDocLen,k1,b float64) float64{
	//this is tf(q,d) q==query d==document
	if tf <= 0 || docLen <= 0 || avgDocLen <= 0 {
		return 0
	}

	tfF := float64(tf)

	norm := k1 * (1.0 - b + b*(float64(docLen)/avgDocLen))


	return (tfF * (k1 + 1.0)) / (tfF + norm)
}
func NewBM25(idx *index.Index) *BM25Engine {
	bm25 := &BM25Engine{
		idx: idx, 
	}
	bm25.defaults()


	return bm25
}


func (bm *BM25Engine) ScoreDoc(queryTerms []string, docID int) float64 {
	
	return 0
}


