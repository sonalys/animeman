package metadata_test

import (
	"testing"

	"github.com/sonalys/animeman/internal/pkg/metadata"
)

func TestTags_String(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "single season",
			input: "S1",
			want:  "S1",
		},
		{
			name:  "season episode",
			input: "S1E13",
			want:  "S1E13",
		},
		{
			name:  "season episode range",
			input: "S1E1-13",
			want:  "S1E1-13",
		},
		{
			name:  "multiple seasons",
			input: "S1 S1 S2 S1 S3E1-13",
			want:  "S1 S2 S3E1-13",
		},
		{
			name:  "multiple seasons",
			input: "S2E01~13",
			want:  "S2E1-13",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := metadata.ParseTags(tt.input)
			got := s.String()

			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
