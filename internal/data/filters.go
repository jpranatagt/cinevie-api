package data

import (
	"math"
	"strings"

	"api.cinevie.jpranata.tech/internal/validator"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

// holding pagination meta-data
type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

func (f Filters) limit() int {
	return f.PageSize
}

// integer overflow was handled by validation
func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

func calculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}

func ValidateFilters(v *validator.Validator, f Filters) {
	// check if page and page_size contain sensible values
	v.Check(f.Page > 0, "page", "must be greater than zero.")
	v.Check(f.Page <= 10_000_000, "page", "must be a maximum of 10 million.")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero.")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100.")

	v.Check(validator.In(f.Sort, f.SortSafelist...), "sort", "invalid sort value.")
}

// if client-provided sort field which matches one of the entries in
// safelist, extract the column name from the sort field by stripping
// the leading hyphen character (if one exists)
// take the Filters struct directly from this file
func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafelist {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-") // trimming the hyphen
		}
	}

	panic("unsafe sort parameter: " + f.Sort)
}

// return the sort direction (ASC or DESC) depends on the prefix character
// of Sort field
func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}

	return "ASC"
}
