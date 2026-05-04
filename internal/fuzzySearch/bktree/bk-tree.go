package bktree

import (
    "github.com/waiyneee/Simplesearch/internal/fuzzySearch/engine"
	"github.com/waiyneee/Simplesearch/internal/fuzzySearch/levenshtein"
)

type BKTree interface {
	Build(words []string) //building atree from a word list
	//for n edits computation

	Add(word string) //adds or inserts  single word

	Search(query string, maxDistance int, candidates []string, limit int) []engine.Suggestion
	// Search returns matches within maxDist, optionally filtered by candidates.

}


//simple tree structre usage in Go

type Node struct {
	Word     string
	Children map[int]*Node
}

type Tree struct {
	Root *Node
}

func New() *Tree {
	return &Tree{}
}

// Build resets the tree and inserts all words.
func (t *Tree) Build(words []string) {
	t.Root = nil
	for _, w := range words {
		t.Add(w)
	}
}


func (t *Tree) Add(word string) {
	if t.Root == nil {
		t.Root = &Node{
			Word:     word,
			Children: make(map[int]*Node),
		}
		return
	}
	t.Root.addRecursive(word)
}

func (n *Node) addRecursive(word string) {
	distance := levenshtein.Compute(n.Word, word)
	if distance == 0 {
		return
	}

	if child, exists := n.Children[distance]; exists {
		child.addRecursive(word)
	} else {
		n.Children[distance] = &Node{
			Word:     word,
			Children: make(map[int]*Node),
		}
	}
}

// Search finds all words within maxDistance edits.
// If candidates is provided, results are filtered to only those words.
func (t *Tree) Search(query string, maxDistance int, candidates []string, limit int) []engine.Suggestion {
	if t.Root == nil {
		return nil
	}

	// Build candidate set for quick filtering (optional).
	var candidateSet map[string]struct{}
	if len(candidates) > 0 {
		candidateSet = make(map[string]struct{}, len(candidates))
		for _, w := range candidates {
			candidateSet[w] = struct{}{}
		}
	}

	results := []engine.Suggestion{}

	var searchNode func(n *Node) bool
	searchNode = func(n *Node) bool {
		if n == nil {
			return false
		}

		d := levenshtein.Compute(n.Word, query)
		if d <= maxDistance {
			if candidateSet == nil || has(candidateSet, n.Word) {
				results = append(results, engine.Suggestion{
					Word:     n.Word,
					Distance: d,
					Score:    scoreFromDistance(d),
				})
				if limit > 0 && len(results) >= limit {
					return true // stop search
				}
			}
		}

		// Only explore children whose edge distance is in [d-maxDistance, d+maxDistance]
		low := d - maxDistance
		high := d + maxDistance

		for dist, child := range n.Children {
			if dist >= low && dist <= high {
				if searchNode(child) {
					return true
				}
			}
		}
		return false
	}

	searchNode(t.Root)
	return results
}

func scoreFromDistance(d int) float64 {
	return 1.0 / float64(d+1)
	//simple scoring will opimize 
	//futher in future usecase 
}

func has(set map[string]struct{}, word string) bool {
	_, ok := set[word]
	return ok
}