package universal

import (
	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("", New())
}

type extractor struct{}

// New returns a universal extractor.
func New() extractors.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	filename, ext, err := utils.GetNameAndExt(url)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	size, err := request.Size(url, url)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	streams := map[string]*extractors.Stream{
		"default": {
			Parts: []*extractors.Part{
				{
					URL:  url,
					Size: size,
					Ext:  ext,
				},
			},
			Size: size,
		},
	}
	contentType, err := request.ContentType(url, url)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return []*extractors.Data{
		{
			Site:    "Universal",
			Title:   filename,
			Type:    extractors.DataType(contentType),
			Streams: streams,
			URL:     url,
		},
	}, nil
}
