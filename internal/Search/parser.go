package search

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	maxQueryBytes = 4096
	maxQueryTerms = 64
)

var (
	ErrEmptyQuery    = errors.New("query is empty")
	ErrQueryTooLong  = fmt.Errorf("query exceeds %d bytes", maxQueryBytes)
	ErrTooManyTerms  = fmt.Errorf("query exceeds %d terms", maxQueryTerms)
	ErrInvalidUTF8   = errors.New("query contains invalid UTF-8")
	ErrInvalidTopK   = errors.New("k must be > 0")
	ErrEngineNotInit = errors.New("search engine not initialized")
)

// ParsedQuery is the validated/normalized query representation.
type ParsedQuery struct {
	Raw   string
	Terms []string
}

// ParseAndValidateQuery performs strict query validation + normalization.
func ParseAndValidateQuery(raw string) (ParsedQuery, error) {
	if !utf8.ValidString(raw) {
		return ParsedQuery{}, ErrInvalidUTF8
	}

	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ParsedQuery{}, ErrEmptyQuery
	}

	if len(trimmed) > maxQueryBytes {
		return ParsedQuery{}, ErrQueryTooLong
	}

	terms := NormalizeQuery(trimmed)
	if len(terms) == 0 {
		return ParsedQuery{}, ErrEmptyQuery
	}
	if len(terms) > maxQueryTerms {
		return ParsedQuery{}, ErrTooManyTerms
	}

	return ParsedQuery{
		Raw:   trimmed,
		Terms: terms,
	}, nil
}
