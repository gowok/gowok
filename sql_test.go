package gowok

import (
	"errors"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang-must/must"
	"github.com/gowok/gowok/config"
	"github.com/ngamux/ngamux"
)

func TestSQL_Conn(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer func() { _ = db.Close() }()

	t.Run("positive/Conn default", func(t *testing.T) {
		p := &_sql{sqls: &sync.Map{}}
		p.sqls.Store("default", db)

		res := p.Conn("default")
		must.True(t, res.IsPresent())
		must.Equal(t, db, res.OrElse(nil))
	})

	t.Run("positive/Conn fallback to default", func(t *testing.T) {
		p := &_sql{sqls: &sync.Map{}}
		p.sqls.Store("default", db)

		res := p.Conn("other")
		must.True(t, res.IsPresent())
		must.Equal(t, db, res.OrElse(nil))
	})

	t.Run("negative/Conn not found", func(t *testing.T) {
		p := &_sql{sqls: &sync.Map{}}
		res := p.Conn("default")
		must.False(t, res.IsPresent())
	})
}

func TestSQL_ConnNoDefault(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer func() { _ = db.Close() }()

	t.Run("positive/found", func(t *testing.T) {
		p := &_sql{sqls: &sync.Map{}}
		p.sqls.Store("other", db)

		res := p.ConnNoDefault("other")
		must.True(t, res.IsPresent())
		must.Equal(t, db, res.OrElse(nil))
	})

	t.Run("negative/not found", func(t *testing.T) {
		p := &_sql{sqls: &sync.Map{}}
		res := p.ConnNoDefault("other")
		must.False(t, res.IsPresent())
	})
}

func TestSQL_healthFunc(t *testing.T) {
	db, mock, _ := sqlmock.New(sqlmock.MonitorPingsOption(true))
	defer func() { _ = db.Close() }()
	p := &_sql{}
	hf := p.healthFunc(db)

	t.Run("positive/UP", func(t *testing.T) {
		mock.ExpectPing()
		res := hf().(ngamux.Map)
		must.Equal(t, "UP", res["status"])
		must.Nil(t, mock.ExpectationsWereMet())
	})

	t.Run("negative/DOWN", func(t *testing.T) {
		mock.ExpectPing().WillReturnError(errors.New("ping failed"))
		res := hf().(ngamux.Map)
		must.Equal(t, "DOWN", res["status"])
		must.Nil(t, mock.ExpectationsWereMet())
	})
}

func TestSQL_configure(t *testing.T) {
	t.Run("positive/nothing enabled", func(t *testing.T) {
		p := &_sql{
			sqls:    &sync.Map{},
			drivers: map[string][]string{"postgres": {"postgres"}},
			plugin:  "sql",
		}
		p.configure(map[string]config.SQL{
			"default": {Enabled: false},
		})
		must.False(t, p.ConnNoDefault("default").IsPresent())
	})

	t.Run("negative/unknown driver", func(t *testing.T) {
		p := &_sql{
			sqls:    &sync.Map{},
			drivers: map[string][]string{"postgres": {"postgres"}},
			plugin:  "sql",
		}
		p.configure(map[string]config.SQL{
			"default": {Enabled: true, Driver: "unknown"},
		})
		must.False(t, p.ConnNoDefault("default").IsPresent())
	})
}
