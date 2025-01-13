package data

import "time"

type Entity struct {
	ID        string     `json:"id,omitempty"`
	CreatedAt time.Time  `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type Health struct {
  Status string `json:"status"`
  Databases map[string]string `json:"databases,omitempty"`
}
