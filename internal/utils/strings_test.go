package utils

import (
	"testing"
)

func TestMatchPrefixFlexible(t *testing.T) {
	tests := []struct {
		name        string // Description of the test case
		s           string // The full string
		prefix      string // The prefix to check
		ignoreChars string // Characters to skip
		want        bool   // Expected result
	}{
		// --- Basic Functionality ---
		{
			name:        "Exact match",
			s:           "Hello World",
			prefix:      "Hello",
			ignoreChars: "",
			want:        true,
		},
		{
			name:        "Case insensitive match",
			s:           "Hello World",
			prefix:      "hELLo",
			ignoreChars: "",
			want:        true,
		},
		{
			name:        "No match",
			s:           "Hello World",
			prefix:      "World",
			ignoreChars: "",
			want:        false,
		},

		// --- Ignoring Characters ---
		{
			name:        "Ignore chars in Source",
			s:           "P-h-o-n-e",
			prefix:      "Phone",
			ignoreChars: "-",
			want:        true,
		},
		{
			name:        "Ignore chars in Source",
			s:           "P-h-x-n-e",
			prefix:      "Phone",
			ignoreChars: "-",
			want:        false,
		},
		{
			name:        "Ignore chars in Prefix",
			s:           "Phone",
			prefix:      "P_h_o",
			ignoreChars: "_",
			want:        true,
		},
		{
			name:        "Ignore chars in Both",
			s:           "123-456",
			prefix:      "123_4",
			ignoreChars: "-_",
			want:        true,
		},
		{
			name:        "Ignore spaces",
			s:           "New York City",
			prefix:      "newyork",
			ignoreChars: " ",
			want:        true,
		},

		// --- Edge Cases ---
		{
			name:        "Prefix longer than Source (Clean)",
			s:           "Short",
			prefix:      "ShortLong",
			ignoreChars: "",
			want:        false,
		},
		{
			name:        "Prefix longer than Source (After ignore)",
			s:           "A-B",
			prefix:      "ABC",
			ignoreChars: "-",
			want:        false,
		},
		{
			name:        "Empty Prefix (Always True)",
			s:           "Anything",
			prefix:      "",
			ignoreChars: "",
			want:        true,
		},
		{
			name:        "Empty Source (False unless prefix empty)",
			s:           "",
			prefix:      "ABC",
			ignoreChars: "",
			want:        false,
		},
		{
			name:        "Both Empty",
			s:           "",
			prefix:      "",
			ignoreChars: "",
			want:        true,
		},
		{
			name:        "Source contains ONLY ignored chars",
			s:           "---",
			prefix:      "A",
			ignoreChars: "-",
			want:        false,
		},

		// --- Unicode / UTF-8 ---
		{
			name:        "Unicode Case Insensitivity (Cyrillic)",
			s:           "Привет", // Privet
			prefix:      "при",    // pri
			ignoreChars: "",
			want:        true,
		},
		{
			name:        "Unicode with ignored chars",
			s:           "GmbH & Co",
			prefix:      "gmbhco",
			ignoreChars: " &",
			want:        true,
		},
		{
			name:        "Emoji handling",
			s:           "🍎🍊🍇",
			prefix:      "🍎🍊",
			ignoreChars: "",
			want:        true,
		},
		{
			name:        "Ignored Emoji",
			s:           "Go🚀Fast",
			prefix:      "gofast",
			ignoreChars: "🚀",
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchPrefixFlexible(tt.s, tt.prefix, tt.ignoreChars)
			if got != tt.want {
				t.Errorf("MatchPrefixFlexible(%q, %q, %q) = %v; want %v",
					tt.s, tt.prefix, tt.ignoreChars, got, tt.want)
			}
		})
	}
}
