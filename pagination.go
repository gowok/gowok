package gowok

import (
	"encoding/json"
	"math"
	"net/http"

	"github.com/ngamux/ngamux"
)

type Pagination[T any] struct {
	Page        int `query:"page" json:"page"`
	PerPage     int `query:"per_page" json:"per_page"`
	TotalRecord int `json:"total_record"`

	Filter map[string]any    `query:"filter" json:"filter"`
	Fields []string          `query:"fields" json:"fields"`
	Sort   map[string]string `query:"sort" json:"sort"`

	Data []T `json:"data"`
}

func PaginationFromReq[T any](r *http.Request) Pagination[T] {
	req := ngamux.Req(r)
	pagination := Pagination[T]{
		Page:    1,
		PerPage: 10,
		Filter:  map[string]any{},
		Sort:    map[string]string{},
	}
	err := req.QueriesParser(&pagination)
	_ = err

	sortQ := req.Query("sort", "{}")
	err = json.Unmarshal([]byte(sortQ), &pagination.Sort)
	_ = err

	filterQ := req.Query("filter", "{}")
	err = json.Unmarshal([]byte(filterQ), &pagination.Filter)
	_ = err

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
