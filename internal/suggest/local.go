package suggest

import (
	"github.com/waiyneee/Simplesearch/internal/fuzzySearch/engine"
	"github.com/waiyneee/Simplesearch/internal/index"
)

type LocalIndexSuggestor struct {
	fuzzy *engine.EngineImpl
}

func NewLocalIndexSuggestor(idx *index.Index) *LocalIndexSuggestor {
	if idx == nil {
		return &LocalIndexSuggestor{fuzzy: nil}
	}
	return &LocalIndexSuggestor{
		fuzzy: engine.New(idx.Terms()),
	}
}

func (s *LocalIndexSuggestor) Suggest(query string, levDist int, trigramSim float64) (string, bool) {
	if s == nil || s.fuzzy == nil {
		return "", false
	}
	// ignore levDist/trigramSim for now; fuzzy engine uses its own tuning
	return s.fuzzy.BestCorrection(query)
}