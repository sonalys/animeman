package dtos

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type SeasonMetadata struct {
	Season string `json:"season,omitempty"`
	Year   int    `json:"year,omitempty"`
	Month  int    `json:"month,omitempty"`
	Day    int    `json:"day,omitempty"`
}

func (t SeasonMetadata) Value() (driver.Value, error) {
	return json.Marshal(t)
}

func (t *SeasonMetadata) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &t)
}
