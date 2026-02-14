package domain

import (
	"time"
)

type Chapter struct {
	Index     uint
	Title     string
	StartTime time.Duration
	EndTime   time.Duration
}
