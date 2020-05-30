package lib

import (
	"github.com/picatz/homoglyphr"
)

// GetHomoglyphMap generates a map of homoglyphs for replacement
func GetHomoglyphMap() map[string]string {
	alphabet := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}
	homoglyph := map[string]string{}
	for _, letter := range alphabet {
		for i := range homoglyphr.StreamAllRelatedCharacters(letter) {
			homoglyph[i] = letter
		}
	}
	return homoglyph
}

// replaceHomoglyph replaces homoglyphs in a string by close latin letters
func replaceHomoglyph(idn string, homoglyphMap map[string]string) string {
	var s string
	for _, i := range idn {
		if i > 127 {
			if letter, present := homoglyphMap[string(i)]; present {
				s += letter
			} else {
				s += string(i)
			}
		} else {
			s += string(i)
		}
	}
	return s
}
