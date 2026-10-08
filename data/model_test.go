package data

import (
	"testing"
	"time"

	"github.com/golang-must/must"
	"github.com/google/uuid"
	"github.com/gowok/gowok/sql"
)

func TestModelFromEntity(t *testing.T) {
	now := time.Now()
	id := uuid.New().String()

	testCases := []struct {
		name   string
		entity Entity
	}{
		{
			name: "positive/full entity",
			entity: Entity{
				ID:        id,
				CreatedAt: now,
				UpdatedAt: &now,
				DeletedAt: &now,
			},
		},
		{
			name: "positive/entity with no optional fields",
			entity: Entity{
				ID:        id,
				CreatedAt: now,
			},
		},
		{
			name: "negative/entity with invalid uuid",
			entity: Entity{
				ID:        "invalid-uuid",
				CreatedAt: now,
			},
		},
		{
			name: "negative/entity with empty id",
			entity: Entity{
				ID:        "",
				CreatedAt: now,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := ModelFromEntity(&tc.entity)
			must.Equal(t, tc.entity.CreatedAt, res.CreatedAt)

			expectedUUID, err := uuid.Parse(tc.entity.ID)
			if err == nil {
				must.Equal(t, expectedUUID, res.ID)
			} else {
				must.Equal(t, uuid.NullUUID{}.UUID, res.ID)
			}

			if tc.entity.UpdatedAt != nil {
				must.True(t, res.UpdatedAt.Valid)
				must.Equal(t, *tc.entity.UpdatedAt, res.UpdatedAt.V)
			} else {
				must.False(t, res.UpdatedAt.Valid)
			}

			if tc.entity.DeletedAt != nil {
				must.True(t, res.DeletedAt.Valid)
				must.Equal(t, *tc.entity.DeletedAt, res.DeletedAt.V)
			} else {
				must.False(t, res.DeletedAt.Valid)
			}
		})
	}
}

func TestModel_Entity(t *testing.T) {
	now := time.Now()
	id := uuid.New()

	testCases := []struct {
		name  string
		model Model
	}{
		{
			name: "positive/full model",
			model: Model{
				ID:        id,
				CreatedAt: now,
				UpdatedAt: sql.NewNull(&now),
				DeletedAt: sql.NewNull(&now),
			},
		},
		{
			name: "positive/model with no optional fields",
			model: Model{
				ID:        id,
				CreatedAt: now,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.model.Entity()
			must.Equal(t, tc.model.ID.String(), res.ID)
			must.Equal(t, tc.model.CreatedAt, res.CreatedAt)

			if tc.model.UpdatedAt.Valid {
				must.NotNil(t, res.UpdatedAt)
				must.Equal(t, tc.model.UpdatedAt.V, *res.UpdatedAt)
			} else {
				must.Nil(t, res.UpdatedAt)
			}

			if tc.model.DeletedAt.Valid {
				must.NotNil(t, res.DeletedAt)
				must.Equal(t, tc.model.DeletedAt.V, *res.DeletedAt)
			} else {
				must.Nil(t, res.DeletedAt)
			}
		})
	}
}

func TestModel_BeforeCreateID_UUID(t *testing.T) {
	testCases := []struct {
		name      string
		model     Model
		expectNew bool
	}{
		{
			name:      "positive/empty id generates new",
			model:     Model{},
			expectNew: true,
		},
		{
			name: "positive/existing id remains",
			model: Model{
				ID: uuid.New(),
			},
			expectNew: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			originalID := tc.model.ID
			err := tc.model.BeforeCreateID_UUID()
			must.Nil(t, err)

			if tc.expectNew {
				must.NotEqual(t, uuid.UUID{}, tc.model.ID)
			} else {
				must.Equal(t, originalID, tc.model.ID)
			}
		})
	}
}
