package facebook

import (
	"fmt"

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
	titles := utils.MatchOneOf(html, `<title id="pageTitle">(.+)</title>`)
	if titles == nil || len(titles) < 2 {
		return nil, types.ErrURLParseFailed
	}
	title := titles[1]

	streams := make(map[string]*types.Stream, 2)
	for _, quality := range []string{"sd", "hd"} {
		srcElement := utils.MatchOneOf(
			html, fmt.Sprintf(`%s_src_no_ratelimit:"(.+?)"`, quality),
		)
		if srcElement == nil || len(srcElement) < 2 {
			continue
		}

		u := srcElement[1]
		size, err := request.Size(u, url)
		if err != nil {
			return nil, err
		}
		urlData := &types.Part{
			URL:  u,
			Size: size,
			Ext:  "mp4",
		}
		streams[quality] = &types.Stream{
			Parts:   []*types.Part{urlData},
			Size:    size,
			Quality: quality,
		}
	}

	return []*types.Data{
		{
			Site:    "Facebook facebook.com",
			Title:   title,
			Type:    types.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
