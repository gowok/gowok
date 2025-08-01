package sql

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/ngamux/ngamux"
)

func NewNullString(inp *string) sql.NullString {
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

func NewNullTime(inp *time.Time) NullTime {
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

func NewNullUUID(inp *uuid.UUID) uuid.NullUUID {
	res := uuid.NullUUID{}
	if inp != nil {
		res.Valid = true
		res.UUID = *inp
	}

	return res
}

type JsonB ngamux.Map

func (j JsonB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JsonB) Scan(value any) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}
