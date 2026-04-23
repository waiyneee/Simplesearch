package app

import (
	"github.com/waiyneee/Simplesearch/internal/Search"
	"github.com/waiyneee/Simplesearch/internal/index"

	"errors"
)

var (
	ErrNilIndex    = errors.New("nil index")
	ErrEmptyQuery  = errors.New("query cannot be empty")
	ErrInvalidTopK = errors.New("topK must be > 0")
	ErrEngineInit  = errors.New("failed to initialize search engine")
)

type App struct {
	idx    *index.Index
	engine *search.Engine
}

func New(idx *index.Index) (*App, error) {
	if idx == nil {
		return nil, ErrNilIndex
	}

	engine := search.NewEngine(idx)
	if engine == nil {
		return nil, ErrEngineInit
	}

	return &App{
		idx:    idx,
		engine: engine,
	}, nil
}
