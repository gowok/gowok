package gowok

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-must/must"
	"github.com/ngamux/ngamux"
)

func TestHealth_Add(t *testing.T) {
	Health.healths = make(map[string]func() any)

	t.Run("positive/add valid health check", func(t *testing.T) {
		hf := func() any { return "ok" }
		Health.Add("test", hf)
		must.Equal(t, 1, len(Health.healths))
		must.NotNil(t, Health.healths["test"])
	})

	t.Run("negative/add empty name", func(t *testing.T) {
		Health.Add("", func() any { return "ok" })
		must.Equal(t, 1, len(Health.healths))
	})

	t.Run("negative/add nil function", func(t *testing.T) {
		Health.Add("nil", nil)
		must.Equal(t, 1, len(Health.healths))
	})
}

func TestHealth_Configure(t *testing.T) {
	oldMux := Web.HttpServeMux
	defer func() { Web.HttpServeMux = oldMux }()
	Web.HttpServeMux = ngamux.NewHttpServeMux()

	Health.healths = make(map[string]func() any)
	Health.Add("ping", func() any { return ngamux.Map{"status": "pong"} })
	Health.Configure()

	t.Run("positive/get health list", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/health", nil)
		Web.ServeHTTP(w, r)

		must.Equal(t, http.StatusOK, w.Code)
		var res map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &res)
		must.Nil(t, err)
		must.Equal(t, "/health/ping", res["ping"])
	})

	t.Run("positive/get specific health", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/health/ping", nil)
		Web.ServeHTTP(w, r)

		must.Equal(t, http.StatusOK, w.Code)
		var res map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &res)
		must.Nil(t, err)
		must.Equal(t, "pong", res["status"])
	})

	t.Run("negative/get non-existent health", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/health/unknown", nil)
		Web.ServeHTTP(w, r)

		must.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("negative/get empty name in path", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/health/%20", nil)
		Web.ServeHTTP(w, r)

		must.Equal(t, http.StatusNotFound, w.Code)
	})
}
