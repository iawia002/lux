package types

import (
	"errors"
)

var (
	// ErrURLParseFailed defines url parse failed error.
	ErrURLParseFailed = errors.New("url parse failed")
	// ErrLoginRequired defines login required error.
	ErrLoginRequired = errors.New("login required")
)
