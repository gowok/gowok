package data

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Model struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt NullTime  `json:"updated_at"`
	DeletedAt NullTime  `json:"deleted_at"`
}

func ModelFromEntity(e *Entity) Model {
	res := Model{
		ID:        uuid.NullUUID{}.UUID,
		CreatedAt: e.CreatedAt,
		UpdatedAt: SQLNullTime(e.UpdatedAt),
		DeletedAt: SQLNullTime(e.DeletedAt),
	}

	if id, err := uuid.Parse(e.ID); err == nil {
		res.ID = id
	}

	return res
}

func (m Model) Entity() Entity {
	res := Entity{
		ID:        m.ID.String(),
		CreatedAt: m.CreatedAt,
	}

	if m.UpdatedAt.Valid {
		res.UpdatedAt = &m.UpdatedAt.Time
	}
	if m.DeletedAt.Valid {
		res.DeletedAt = &m.DeletedAt.Time
	}

	return res
}

func SQLNullString(inp *string) sql.NullString {
	res := sql.NullString{}
	if inp != nil {
		res.Valid = true
		res.String = *inp
	}

	return res
}

type NullTime struct {
	sql.NullTime
}

func SQLNullTime(inp *time.Time) NullTime {
	res := sql.NullTime{}
	if inp != nil {
		res.Valid = true
		res.Time = *inp
	}

	return NullTime{res}
}

func (nt NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(nt.Time)
}

func (nt *NullTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		nt.Valid = false
		return nil
	}

	var t time.Time
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}

	nt.Time = t
	nt.Valid = true
	return nil
}

func SQLNullUUID(inp *uuid.UUID) uuid.NullUUID {
	res := uuid.NullUUID{}
	if inp != nil {
		res.Valid = true
		res.UUID = *inp
	}

	return res
}
