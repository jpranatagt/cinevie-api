package data

import (
  "time"
)

// anotate the Movie struct with struct tags
// to control how they appear in the JSON encoded output
// use snake_case for the keys instead of CamelCase
// add directive "-" to hide a field and "omitempty" if only if it's empty
type Movie struct {
  ID          int64       `json:"id"`
  CreatedAt   time.Time   `json:"-"`
  Title       string      `json:"title"`
  Description string      `json:"description,omitempty"`
  Cover       string      `json:"cover,omitempty"`
  Trailer     string      `json:"trailer,omitempty"`
  Year        int32       `json:"year,omitempty"`
  Runtime     int32       `json:"runtime,omitempty"`
  Genres      []string    `json:"genres,omitempty"`
  Cast        []string    `json:"cast,omitempty"`
  Version     int32       `json:"version"`
}


