package javdbapi

import "encoding/json"

// Page never claims a total page count: the site only exposes rel="next" pagination.
type Page[T any] struct {
	Items   []T  `json:"items"`
	Number  int  `json:"number"`
	HasNext bool `json:"has_next"`
}

func NewPage[T any](number int, items []T, hasNext bool) Page[T] {
	if items == nil {
		items = []T{}
	}
	return Page[T]{Items: items, Number: number, HasNext: hasNext}
}

func (p Page[T]) MarshalJSON() ([]byte, error) {
	items := p.Items
	if items == nil {
		items = []T{}
	}
	type pageJSON struct {
		Items   []T  `json:"items"`
		Number  int  `json:"number"`
		HasNext bool `json:"has_next"`
	}
	return json.Marshal(pageJSON{Items: items, Number: p.Number, HasNext: p.HasNext})
}
