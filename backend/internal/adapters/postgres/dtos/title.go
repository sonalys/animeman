package dtos

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type (
	Titles []Title

	Title struct {
		TitleValue string `json:"value,omitempty"`
		Language   string `json:"language,omitempty"`
		Type       string `json:"type,omitempty"`
	}
)

func (t Titles) Value() (driver.Value, error) {
	return json.Marshal(t)
}

func (t *Titles) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &t)
}
