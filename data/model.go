package data

import (
	"time"

	"github.com/google/uuid"
	"github.com/gowok/gowok/sql"
)

type Model struct {
	ID        uuid.UUID           `json:"id"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt sql.Null[time.Time] `json:"updated_at"`
	DeletedAt sql.Null[time.Time] `json:"deleted_at"`
}

func ModelFromEntity(e *Entity) Model {
	res := Model{
		ID:        uuid.NullUUID{}.UUID,
		CreatedAt: e.CreatedAt,
		UpdatedAt: sql.NewNull(e.UpdatedAt),
		DeletedAt: sql.NewNull(e.DeletedAt),
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
		res.UpdatedAt = &m.UpdatedAt.V
	}
	if m.DeletedAt.Valid {
		res.DeletedAt = &m.DeletedAt.V
	}

	return res
}

func (m *Model) BeforeCreateID_UUID() error {
	if m.ID == (uuid.NullUUID{}).UUID {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		m.ID = id
	}
	return nil
}
