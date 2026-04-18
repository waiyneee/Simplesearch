package index

import (
	"strings"
	"unicode"
)

func NormalizeText(text string) string {
	text = strings.ToLower(text)
	text = strings.TrimSpace(text)

	var cleaned []rune

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			cleaned = append(cleaned, r)
		} else {
			// replace punctuation with space
			cleaned = append(cleaned, ' ')
		}
	}

	
	return strings.Join(strings.Fields(string(cleaned)), " ")
}


func Tokenize(text string) []string{
	//tokenizer is deterministic and wont fail so no need of errors 
	cleanedText:=NormalizeText(text)
	if cleanedText==""{
		return nil
	}
	tokens:=strings.Fields(cleanedText) //return  a slice 



    //will add optional linguistic oprocessing 
	// optional words "this" //stemming as well lateron


	return tokens

}