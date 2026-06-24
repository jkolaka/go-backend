package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type JSONSlice []string

func (s JSONSlice) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "[]", nil
	}
	return json.Marshal(s)
}

func (s *JSONSlice) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, s)
}

type Recipe struct {
	ID           string    `gorm:"primaryKey" json:"id"`
	Name         string    `json:"name"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Tags         JSONSlice `gorm:"type:text" json:"tags"`
	Ingredients  JSONSlice `gorm:"type:text" json:"ingredients"`
	Instructions JSONSlice `gorm:"type:text" json:"instructions"`
	PublishedAt  time.Time `json:"publishedAt"`
	UserID       string    `json:"userId" gorm:"index"`
	User         User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
}