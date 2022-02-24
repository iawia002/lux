package extractors

import (
	"github.com/pkg/errors"
)

var (
	// ErrURLParseFailed defines url parse failed error.
	ErrURLParseFailed           = errors.New("url parse failed")
	ErrInvalidRegularExpression = errors.New("invalid regular expression")
)
