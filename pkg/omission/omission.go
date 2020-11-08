package omission

import (
	"cercat/helper"
	"fmt"
)

// GetOmissionPatterns returns a list of strings with missing letters
func GetOmissionPatterns(domain string) []string {
	results := []string{}
	for i := range domain {
		results = append(results, fmt.Sprintf("%s%s", domain[:i], domain[i+1:]))
	}
	return helper.RemoveDuplicate(results)
}
