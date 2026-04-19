package index

import (
	"strings"
	"unicode"
)

func NormalizeText(text string) string {
	text = strings.ToLower(strings.TrimSpace(text))

	var cleaned []rune
	cleaned = make([]rune, 0, len(text))

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {

			
			cleaned = append(cleaned, r)
		} else {
			// Replace punctuation/symbols with space
			cleaned = append(cleaned, ' ')
		}
	}

	// Collapse consecutive whitespace
	return strings.Join(strings.Fields(string(cleaned)), " ")
}

func Tokenize(text string) []string {
	cleaned := NormalizeText(text)
	if cleaned == "" {
		return []string{}
	}
	return strings.Fields(cleaned)
}