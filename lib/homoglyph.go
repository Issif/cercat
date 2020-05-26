package lib

import (
	"github.com/picatz/homoglyphr"
)

func getHomoglyphMap() map[string]string {
	alphabet := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}
	homoglyph := map[string]string{}
	for _, letter := range alphabet {
		for i := range homoglyphr.StreamAllRelatedCharacters(letter) {
			homoglyph[i] = letter
		}
	}
	return homoglyph
}
