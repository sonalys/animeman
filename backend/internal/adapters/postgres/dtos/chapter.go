package dtos

import "time"

type (
	Chapters []Chapter

	Chapter struct {
		Title     string        `json:"title,omitzero"`
		StartTime time.Duration `json:"start_time,omitzero"`
		EndTime   time.Duration `json:"end_time,omitzero"`
	}
)
