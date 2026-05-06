package suggest

import "strings"

type LocalSuggestor interface {
	Suggest(query string, levDist int, trigramSim float64) (string, bool)
}

type WikiSuggestor interface {
	TopTitle(query string) (string, error)
}

type Corrector struct {
	cache      Cache
	local      LocalSuggestor
	wiki       WikiSuggestor
	levDist    int
	trigramSim float64
}

func NewCorrector(cache Cache, local LocalSuggestor, wiki WikiSuggestor) *Corrector {
	if cache == nil {
		cache = NewMemoryCache()
	}
	return &Corrector{
		cache:      cache,
		local:      local,
		wiki:       wiki,
		levDist:    2,
		trigramSim: 0.35,
	}
}

func (c *Corrector) Correct(query string) (string, string) {
	q := strings.TrimSpace(query)
	if q == "" {
		return "", "none"
	}

	key := "correct:" + strings.ToLower(q)
	if c.cache != nil {
		if val, ok := c.cache.Get(key); ok && val != "" {
			return val, "cache"
		}
	}

	if c.local != nil {
		if val, ok := c.local.Suggest(q, c.levDist, c.trigramSim); ok && val != "" {
			c.cache.Set(key, val)
			return val, "local"
		}
	}

	if c.wiki != nil {
		if val, err := c.wiki.TopTitle(q); err == nil && val != "" {
			c.cache.Set(key, val)
			return val, "wiki"
		}
	}

	return q, "none"
}