package transposition

import (
	"cercat/helper"
	"fmt"
)

// GetTranspositionPatterns returns a list of strings with swapped characters
func GetTranspositionPatterns(domain string) []string {
	results := []string{}
	for i := range domain[:len(domain)-1] {
		if domain[i+1] != domain[i] {
			results = append(results, fmt.Sprintf("%s%c%c%s", domain[:i], domain[i+1], domain[i], domain[i+2:]))
		}
	}
	return helper.RemoveDuplicate(results)
}
