package extractors

import (
	"errors"
)

var ErrURLParseFailed = errors.New("url parse failed")
var ErrLoginRequired = errors.New("login required")
