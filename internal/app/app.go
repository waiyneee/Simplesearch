package app

import (
	"errors"
	"strings"
	"time"

	"github.com/waiyneee/Simplesearch/internal/Search"
	"github.com/waiyneee/Simplesearch/internal/index"
	"github.com/waiyneee/Simplesearch/internal/suggest"
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
	cfg    Config
}

func New(idx *index.Index, cfg Config) (*App, error) {
	if idx == nil {
		return nil, ErrNilIndex
	}

	cache := buildCache(cfg)
	local := suggest.NewLocalIndexSuggestor(idx)
	wiki := suggest.NewWikiSuggestor()

	corrector := suggest.NewCorrector(cache, local, wiki)

	engine := search.NewEngine(idx, corrector)
	if engine == nil {
		return nil, ErrEngineInit
	}

	return &App{
		idx:    idx,
		engine: engine,
		cfg:    cfg,
	}, nil
}

func buildCache(cfg Config) suggest.Cache {
	mode := strings.ToLower(strings.TrimSpace(cfg.CacheMode))

	if mode == "redis" && cfg.RedisURL != "" {
		redisCache, err := suggest.NewRedisCache(cfg.RedisURL, 24*time.Hour)
		if err == nil {
			return redisCache
		}
	}

	return suggest.NewMemoryCache()
}