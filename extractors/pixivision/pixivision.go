package pixivision

import (
	"github.com/iawia002/annie/extractors/types"
	"github.com/iawia002/annie/parser"
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
	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, err
	}
	title, urls, err := parser.GetImages(html, "am__work__illust  ", nil)
	if err != nil {
		return nil, err
	}

	parts := make([]*types.Part, 0, len(urls))
	for _, u := range urls {
		_, ext, err := utils.GetNameAndExt(u)
		if err != nil {
			return nil, err
		}
		size, err := request.Size(u, url)
		if err != nil {
			return nil, err
		}
		parts = append(parts, &types.Part{
			URL:  u,
			Size: size,
			Ext:  ext,
		})
	}

	streams := map[string]*types.Stream{
		"default": {
			Parts: parts,
		},
	}

	return []*types.Data{
		{
			Site:    "pixivision pixivision.net",
			Title:   title,
			Type:    types.DataTypeImage,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
