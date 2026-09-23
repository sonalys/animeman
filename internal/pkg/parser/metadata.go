package parser

import "github.com/sonalys/animeman/internal/pkg/tags"

// Metadata is a digested metadata struct parsed from titles.
type Metadata struct {
	ReleaseGroup       string
	ShowTitle          string
	Tag                tags.Tag
	Labels             []string
	VerticalResolution int
}

func (m Metadata) Clone() Metadata {
	return Metadata{
		ReleaseGroup:       m.ReleaseGroup,
		ShowTitle:          m.ShowTitle,
		Tag:                m.Tag,
		Labels:             append([]string{}, m.Labels...),
		VerticalResolution: m.VerticalResolution,
	}
}
