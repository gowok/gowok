package sql

import (
	"database/sql"
	"encoding/json"
)

type Null[T any] struct {
	sql.Null[T]
}

func NewNull[T any](inp *T) Null[T] {
	res := sql.Null[T]{}
	if inp != nil {
		res.Valid = true
		res.V = *inp
	}

	return Null[T]{res}
}

func (nt Null[T]) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return json.Marshal(nil)
	}

	return json.Marshal(nt.V)
}

func (nt *Null[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		nt.Valid = false
		return nil
	}

	var t T
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}

	nt.V = t
	nt.Valid = true
	return nil
}
