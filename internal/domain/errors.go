package domain

import "errors"

var (
	ErrNotFound    = errors.New("not found")
	ErrInvalidData = errors.New("invalid data")
	ErrAPI         = errors.New("API error")
	ErrPersistence = errors.New("persistence error")
)
