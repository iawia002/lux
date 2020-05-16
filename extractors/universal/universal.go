package universal

import (
	"github.com/iawia002/annie/extractors/types"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type extractor struct{}

// New returns a youtube extractor.
func New() types.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option types.Options) ([]*types.Data, error) {
	filename, ext, err := utils.GetNameAndExt(url)
	if err != nil {
		return nil, err
	}
	size, err := request.Size(url, url)
	if err != nil {
		return nil, err
	}
	streams := map[string]*types.Stream{
		"default": {
			Parts: []*types.Part{
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
		return nil, err
	}

	return []*types.Data{
		{
			Site:    "Universal",
			Title:   filename,
			Type:    types.DataType(contentType),
			Streams: streams,
			URL:     url,
		},
	}, nil
}
