package douyin

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
	var err error
	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, err
	}
	var title string
	desc := utils.MatchOneOf(html, `<p class="desc">(.+?)</p>`)
	if desc != nil {
		title = desc[1]
	} else {
		title = "抖音短视频"
	}
	realURLs := utils.MatchOneOf(html, `playAddr: "(.+?)"`)
	if realURLs == nil || len(realURLs) < 2 {
		return nil, types.ErrURLParseFailed
	}
	realURL := realURLs[1]

	size, err := request.Size(realURL, url)
	if err != nil {
		return nil, err
	}
	urlData := &types.Part{
		URL:  realURL,
		Size: size,
		Ext:  "mp4",
	}
	streams := map[string]*types.Stream{
		"default": {
			Parts: []*types.Part{urlData},
			Size:  size,
		},
	}
	return []*types.Data{
		{
			Site:    "抖音 douyin.com",
			Title:   title,
			Type:    types.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
