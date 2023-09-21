package cctv

import (
	"errors"

	"github.com/wujiu2020/lux/extractors/proto"
)

func New() proto.Extractor {
	return &extractor{}
}

type extractor struct{}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string) (*proto.Data, error) {
	return nil, errors.New("not implement")
}
