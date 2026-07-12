package scrape

import "errors"

var (
	ErrEmpty = errors.New("empty result")
	ErrParse = errors.New("parse failed")
)

type Warning struct {
	Field   string
	Message string
}

type Result[T any] struct {
	Value    T
	Warnings []Warning
}
