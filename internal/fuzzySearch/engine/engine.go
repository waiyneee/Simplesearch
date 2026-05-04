package engine

type Suggestion struct {
	Word     string
	Distance int
	Score    float64
}

type Engine interface {
	Suggest(query string, limit int) []Suggestion
	BestCorrection(query string) (string, bool)
	AddWord(word string)
}


