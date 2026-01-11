package utils

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func HasPrefixFold(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	return strings.EqualFold(s[:len(prefix)], prefix)
}

// MatchPrefixFlexible checks if 's' starts with 'prefix'.
// It compares case-insensitively and skips characters present in 'ignoreChars'.
func MatchPrefixFlexible(s, prefix, ignoreChars string) bool {
	sIdx, pIdx := 0, 0

	for pIdx < len(prefix) {
		// 1. Decode the next rune from the prefix
		rPrefix, widthPrefix := utf8.DecodeRuneInString(prefix[pIdx:])

		// If the char in the prefix is in the ignore list, skip it
		if strings.ContainsRune(ignoreChars, rPrefix) {
			pIdx += widthPrefix
			continue
		}

		// 2. Find the next relevant rune in 's'
		var rStr rune
		var widthStr int

		for {
			// If 's' runs out of characters before 'prefix', it's not a match
			if sIdx >= len(s) {
				return false
			}

			rStr, widthStr = utf8.DecodeRuneInString(s[sIdx:])

			// If the char in 's' is in the ignore list, skip it and keep looking
			if strings.ContainsRune(ignoreChars, rStr) {
				sIdx += widthStr
				continue
			}

			// Found a valid character
			break
		}

		// 3. Compare the characters case-insensitively
		if unicode.ToLower(rPrefix) != unicode.ToLower(rStr) {
			return false
		}

		// Advance both pointers
		pIdx += widthPrefix
		sIdx += widthStr
	}

	return true
}
