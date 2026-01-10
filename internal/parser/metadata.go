package parser

import "github.com/sonalys/animeman/internal/tags"

// Metadata is a digested metadata struct parsed from titles.
type Metadata struct {
	Source             string
	Title              string
	Tag                tags.Tag
	Labels             []string
	VerticalResolution int
}

func (m Metadata) Clone() Metadata {
	return Metadata{
		Source:             m.Source,
		Title:              m.Title,
		Tag:                m.Tag,
		Labels:             append([]string{}, m.Labels...),
		VerticalResolution: m.VerticalResolution,
	}
}
