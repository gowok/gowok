package sql

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang-must/must"
)

func TestWithTx(t *testing.T) {
	t.Run("positive/success commits", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		must.Equal(t, nil, err)
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectCommit()

		called := false
		err = WithTx(context.Background(), db, func(tx *sql.Tx) error {
			called = true
			return nil
		})

		must.Equal(t, nil, err)
		must.True(t, called)
		must.Equal(t, nil, mock.ExpectationsWereMet())
	})

	t.Run("negative/callback error rolls back", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		must.Equal(t, nil, err)
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectRollback()

		wantErr := errors.New("boom")
		err = WithTx(context.Background(), db, func(tx *sql.Tx) error {
			return wantErr
		})

		must.Equal(t, wantErr, err)
		must.Equal(t, nil, mock.ExpectationsWereMet())
	})

	t.Run("negative/callback panic rolls back and re-panics", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		must.Equal(t, nil, err)
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectRollback()

		func() {
			defer func() {
				r := recover()
				must.NotNil(t, r)
				must.Equal(t, "boom", r)
			}()

			_ = WithTx(context.Background(), db, func(tx *sql.Tx) error {
				panic("boom")
			})
		}()

		must.Equal(t, nil, mock.ExpectationsWereMet())
	})

	t.Run("negative/begin fails skips callback", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		must.Equal(t, nil, err)
		defer db.Close()

		wantErr := errors.New("boom")
		mock.ExpectBegin().WillReturnError(wantErr)

		called := false
		err = WithTx(context.Background(), db, func(tx *sql.Tx) error {
			called = true
			return nil
		})

		must.False(t, called)
		must.True(t, errors.Is(err, wantErr))
		must.True(t, strings.Contains(err.Error(), "sql: begin transaction"))
		must.Equal(t, nil, mock.ExpectationsWereMet())
	})

	t.Run("positive/callback runs queries on the transaction", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		must.Equal(t, nil, err)
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT id FROM users").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
		mock.ExpectCommit()

		var got int
		err = WithTx(context.Background(), db, func(tx *sql.Tx) error {
			return tx.QueryRowContext(context.Background(), "SELECT id FROM users").Scan(&got)
		})

		must.Equal(t, nil, err)
		must.Equal(t, 7, got)
		must.Equal(t, nil, mock.ExpectationsWereMet())
	})

	t.Run("negative/commit failure wraps error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		must.Equal(t, nil, err)
		defer db.Close()

		wantErr := errors.New("boom")
		mock.ExpectBegin()
		mock.ExpectCommit().WillReturnError(wantErr)

		err = WithTx(context.Background(), db, func(tx *sql.Tx) error {
			return nil
		})

		must.True(t, errors.Is(err, wantErr))
		must.True(t, strings.Contains(err.Error(), "sql: commit transaction"))
		must.Equal(t, nil, mock.ExpectationsWereMet())
	})
}
