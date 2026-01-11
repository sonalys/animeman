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

func CalculateTextSimilarity(s1, s2, ignore string) float64 {
	// 1. Get raw distance
	dist := CalculateDistance(s1, s2, ignore)

	// 2. Get the length of the longer string (after cleaning)
	// (You would need to expose the cleaning logic to do this accurately)
	cleaned1 := cleanAndNormalize(s1, ignore)
	cleaned2 := cleanAndNormalize(s2, ignore)

	maxLen := max(len(cleaned2), len(cleaned1))

	if maxLen == 0 {
		return 1.0 // Both empty = 100% match
	}

	// 3. Calculate percentage
	return 1.0 - (float64(dist) / float64(maxLen))
}

// CalculateDistance calculates the Levenshtein distance between two strings,
// applying case-insensitivity and ignoring specific characters.
func CalculateDistance(s1, s2 string, ignoreChars string) int {
	// 1. Pre-process the strings:
	//    - Convert to []rune to handle multi-byte characters (emojis, kanji, etc.)
	//    - Remove ignored characters
	//    - Normalize case (to lower)
	r1 := cleanAndNormalize(s1, ignoreChars)
	r2 := cleanAndNormalize(s2, ignoreChars)

	// 2. Compute Levenshtein distance
	return levenshtein(r1, r2)
}

// cleanAndNormalize converts string to rune slice, removes ignored chars,
// and lowercases the rest.
func cleanAndNormalize(s, ignore string) []rune {
	// Create a lookup map for ignored characters for O(1) access
	ignoreMap := make(map[rune]struct{})
	for _, r := range ignore {
		ignoreMap[r] = struct{}{}
	}

	var result []rune
	for _, r := range s {
		// specific check: should we ignore this specific rune?
		if _, shouldIgnore := ignoreMap[r]; shouldIgnore {
			continue
		}
		// Case insensitive: append the lowercase version
		result = append(result, unicode.ToLower(r))
	}
	return result
}

// levenshtein calculates the edit distance between two rune slices.
// Optimized to use O(min(m,n)) space instead of a full matrix.
func levenshtein(r1, r2 []rune) int {
	len1, len2 := len(r1), len(r2)

	// Optimization: Ensure r1 is the shorter slice to minimize memory usage
	if len1 > len2 {
		r1, r2 = r2, r1
		len1, len2 = len2, len1
	}

	if len1 == 0 {
		return len2
	}

	// Create two rows for the dynamic programming calculation
	// We only need the previous row and the current row
	prevColumn := make([]int, len2+1)
	currColumn := make([]int, len2+1)

	// Initialize the first row (0 to len2)
	for i := 0; i <= len2; i++ {
		prevColumn[i] = i
	}

	for i := 0; i < len1; i++ {
		currColumn[0] = i + 1

		for j := 0; j < len2; j++ {
			cost := 1
			if r1[i] == r2[j] {
				cost = 0
			}

			// Calculate minimum of:
			// 1. Deletion (currColumn[j] + 1)
			// 2. Insertion (prevColumn[j+1] + 1)
			// 3. Substitution (prevColumn[j] + cost)
			ins := prevColumn[j+1] + 1
			del := currColumn[j] + 1
			sub := prevColumn[j] + cost

			minVal := ins
			if del < minVal {
				minVal = del
			}
			if sub < minVal {
				minVal = sub
			}

			currColumn[j+1] = minVal
		}

		// Swap rows for next iteration
		copy(prevColumn, currColumn)
	}

	return prevColumn[len2]
}
