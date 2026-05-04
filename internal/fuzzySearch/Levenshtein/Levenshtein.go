package levenshtein


import "strings"
func Compute(a, b string) int{
	if a=="" || b=="" {
		return 0
	}

	strings.ToLower(a)
	strings.ToLower(b)


	//rune conversion 
	n1:=len(a)
	n2:=len(b)

	




}