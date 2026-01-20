package web

import (
	"bytes"
	"math"
	"net/http"
	"sync"

	"github.com/gowok/gowok/json"
	"github.com/ngamux/ngamux"
)

type Pagination[T any] struct {
	Filter map[string]any    `json:"filter"`
	Fields []string          `json:"fields"`
	Sort   map[string]string `json:"sort"`

	Data []T `json:"data"`

	Page        int `query:"page" json:"page"`
	PerPage     int `query:"per_page" json:"per_page"`
	TotalRecord int `json:"total_record"`
}

var poolByte = sync.Pool{
	New: func() any {
		return &bytes.Buffer{}
	},
}

func PaginationFromReq[T any](r *http.Request) Pagination[T] {
	req := ngamux.Req(r)
	pagination := Pagination[T]{
		Page:    1,
		PerPage: 10,
		Filter:  map[string]any{},
		Sort:    map[string]string{},
	}
	_ = req.QueriesParser(&pagination)

	buf := poolByte.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		poolByte.Put(buf)
	}()

	_ = json.Unmarshal([]byte(req.Query("sort", "{}")), &pagination.Sort)
	_ = json.Unmarshal([]byte(req.Query("filter", "{}")), &pagination.Filter)

	pagination.Data = make([]T, 0)

	return pagination
}

func PaginationFromPagination[T, U any](input Pagination[U]) Pagination[T] {
	pagination := Pagination[T]{
		Page:        input.Page,
		PerPage:     input.PerPage,
		Filter:      input.Filter,
		Sort:        input.Sort,
		TotalRecord: input.TotalRecord,
		Fields:      input.Fields,
		Data:        make([]T, len(input.Data)),
	}

	return pagination
}

func (p *Pagination[T]) SetData(data ...T) {
	p.Data = data
}

func (p Pagination[T]) Skip() int {
	return (p.Page - 1) * p.PerPage
}

func (p Pagination[T]) TotalPage() float64 {
	return math.Ceil(float64(p.TotalRecord) / float64(p.PerPage))
}

func (p Pagination[T]) MarshalJSON() ([]byte, error) {
	res := map[string]any{
		"page":         p.Page,
		"fields":       p.Fields,
		"per_page":     p.PerPage,
		"total_record": p.TotalRecord,
		"total_page":   1,
		"filter":       p.Filter,
		"sort":         p.Sort,
		"data":         p.Data,
	}

	if p.TotalRecord > 0 && p.PerPage > 0 {
		res["total_page"] = p.TotalPage()
	}

	return json.Marshal(res)
}
