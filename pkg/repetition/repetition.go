package repetition

import (
	"cercat/helper"
	"fmt"
	"unicode"
)

// GetRepetitionPatterns returns a  list of strings with doubled letters
func GetRepetitionPatterns(domain string) []string {
	results := []string{}
	count := make(map[string]int)
	for i, c := range domain {
		if unicode.IsLetter(c) {
			result := fmt.Sprintf("%s%c%c%s", domain[:i], domain[i], domain[i], domain[i+1:])
			count[result]++
		}
	}
	return helper.RemoveDuplicate(results)
}
