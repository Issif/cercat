package bitsquatting

import (
	"cercat/helper"
	"fmt"
)

// GetBitsquattingPatterns returns a list of strings with letters swapped for close chracter in Unicode Table
func GetBitsquattingPatterns(domain string) []string {
	results := []string{}
	masks := []int32{1, 2, 4, 8, 16, 32, 64, 128}

	for i, c := range domain {
		for m := range masks {
			b := rune(int(c) ^ m)
			o := int(b)
			if (o >= 48 && o <= 57) || (o >= 97 && o <= 122) || o == 45 {
				results = append(results, fmt.Sprintf("%s%c%s", domain[:i], b, domain[i+1:]))
			}
		}
	}
	return helper.RemoveDuplicate(results)
}
