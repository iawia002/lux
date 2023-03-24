package extractors

import (
	"errors"
)

var (
	// ErrURLParseFailed defines url parse failed error.
	ErrURLParseFailed            = errors.New("url parse failed")
	ErrInvalidRegularExpression  = errors.New("invalid regular expression")
	ErrURLQueryParamsParseFailed = errors.New("url query params parse failed")
)
