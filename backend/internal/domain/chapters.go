package domain

import (
	"time"
)

type Chapter struct {
	Title     string
	StartTime time.Duration
	EndTime   time.Duration
}
