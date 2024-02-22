package parser

import "testing"

func TestTitleStrip(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{name: "empty", title: "", want: ""},
		{name: "multiple spaces", title: "My     cool   anime", want: "My cool anime"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TitleStrip(tt.title); got != tt.want {
				t.Errorf("TitleStrip() = %v, want %v", got, tt.want)
			}
		})
	}
}
