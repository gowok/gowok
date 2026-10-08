package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-must/must"
)

func TestPagination_PaginationFromReq(t *testing.T) {
	testCases := []struct {
		name        string
		url         string
		wantPage    int
		wantPerPage int
		wantSort    map[string]string
		wantFilter  map[string]any
	}{
		{
			name:        "positive/default values",
			url:         "/",
			wantPage:    1,
			wantPerPage: 10,
			wantSort:    map[string]string{},
			wantFilter:  map[string]any{},
		},
		{
			name:        "positive/custom query params",
			url:         "/?page=2&per_page=20",
			wantPage:    2,
			wantPerPage: 20,
			wantSort:    map[string]string{},
			wantFilter:  map[string]any{},
		},
		{
			name:        "positive/complex sort and filter",
			url:         `/?sort={"name":"asc"}&filter={"status":"active"}`,
			wantPage:    1,
			wantPerPage: 10,
			wantSort:    map[string]string{"name": "asc"},
			wantFilter:  map[string]any{"status": "active"},
		},
		{
			name:        "negative/invalid json in sort and filter",
			url:         `/?sort=invalid&filter=invalid`,
			wantPage:    1,
			wantPerPage: 10,
			wantSort:    map[string]string{},
			wantFilter:  map[string]any{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tc.url, nil)
			p := PaginationFromReq[string](r)

			must.Equal(t, tc.wantPage, p.Page)
			must.Equal(t, tc.wantPerPage, p.PerPage)
			must.Equal(t, tc.wantSort, p.Sort)
			must.Equal(t, tc.wantFilter, p.Filter)
		})
	}
}

func TestPagination_PaginationFromPagination(t *testing.T) {
	t.Run("positive/conversion", func(t *testing.T) {
		input := Pagination[int]{
			Page:        1,
			PerPage:     10,
			Filter:      map[string]any{"foo": "bar"},
			Sort:        map[string]string{"baz": "qux"},
			TotalRecord: 100,
			Fields:      []string{"a", "b"},
			Data:        []int{1, 2, 3},
		}

		output := PaginationFromPagination[string, int](input)

		must.Equal(t, input.Page, output.Page)
		must.Equal(t, input.PerPage, output.PerPage)
		must.Equal(t, input.Filter, output.Filter)
		must.Equal(t, input.Sort, output.Sort)
		must.Equal(t, input.TotalRecord, output.TotalRecord)
		must.Equal(t, input.Fields, output.Fields)
		must.Equal(t, len(input.Data), len(output.Data))
	})
}

func TestPagination_SetData(t *testing.T) {
	t.Run("positive/sets data", func(t *testing.T) {
		p := Pagination[int]{}
		p.SetData(1, 2, 3)
		must.Equal(t, []int{1, 2, 3}, p.Data)
	})
}

func TestPagination_Skip(t *testing.T) {
	testCases := []struct {
		name     string
		page     int
		perPage  int
		wantSkip int
	}{
		{
			name:     "positive/page 1",
			page:     1,
			perPage:  10,
			wantSkip: 0,
		},
		{
			name:     "positive/page 2",
			page:     2,
			perPage:  10,
			wantSkip: 10,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := Pagination[int]{Page: tc.page, PerPage: tc.perPage}
			must.Equal(t, tc.wantSkip, p.Skip())
		})
	}
}

func TestPagination_TotalPage(t *testing.T) {
	testCases := []struct {
		name          string
		totalRecord   int
		perPage       int
		wantTotalPage float64
	}{
		{
			name:          "positive/exact division",
			totalRecord:   100,
			perPage:       10,
			wantTotalPage: 10,
		},
		{
			name:          "positive/with remainder",
			totalRecord:   105,
			perPage:       10,
			wantTotalPage: 11,
		},
		{
			name:          "positive/zero records",
			totalRecord:   0,
			perPage:       10,
			wantTotalPage: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := Pagination[int]{TotalRecord: tc.totalRecord, PerPage: tc.perPage}
			must.Equal(t, tc.wantTotalPage, p.TotalPage())
		})
	}
}

func TestPagination_MarshalJSON(t *testing.T) {
	t.Run("positive/marshal structured output", func(t *testing.T) {
		p := Pagination[int]{
			Page:        1,
			PerPage:     10,
			TotalRecord: 25,
			Data:        []int{1, 2},
		}

		b, err := json.Marshal(p)
		must.Nil(t, err)

		var res map[string]any
		err = json.Unmarshal(b, &res)
		must.Nil(t, err)

		must.Equal(t, float64(1), res["page"])
		must.Equal(t, float64(10), res["per_page"])
		must.Equal(t, float64(25), res["total_record"])
		must.Equal(t, float64(3), res["total_page"])
	})

	t.Run("positive/zero per_page avoids division by zero", func(t *testing.T) {
		p := Pagination[int]{TotalRecord: 10, PerPage: 0}
		b, err := json.Marshal(p)
		must.Nil(t, err)

		var res map[string]any
		err = json.Unmarshal(b, &res)
		must.Nil(t, err)
		must.Equal(t, float64(1), res["total_page"])
	})
}
