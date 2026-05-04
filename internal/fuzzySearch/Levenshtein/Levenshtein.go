package levenshtein

import "strings"

func Compute(a, b string) int {
	a = strings.ToLower(a)
	b = strings.ToLower(b)

	ar := []rune(a) //a rune utf seen
	br := []rune(b) //b rune utf seen

	if len(ar) == 0 {
		return len(br)
	}
	if len(br) == 0 {
		return len(ar)
	}

	dp := make([][]int, len(ar)+1)
	for i := range dp {
		dp[i] = make([]int, len(br)+1)
	}//dp matrix 

	for i := 0; i <= len(ar); i++ {
		dp[i][0] = i
	}
	for j := 0; j <= len(br); j++ {
		dp[0][j] = j
	}

	for i := 1; i <= len(ar); i++ {
		for j := 1; j <= len(br); j++ {
			cost := 0
			if ar[i-1] != br[j-1] {
				cost = 1
			}
			dp[i][j] = min(
				dp[i-1][j]+1,      // delete
				dp[i][j-1]+1,      // insert
				dp[i-1][j-1]+cost, // substitute
			)
		}
	}

	return dp[len(ar)][len(br)]
}

func min(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
}
