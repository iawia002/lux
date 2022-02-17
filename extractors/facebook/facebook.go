package facebook

import (
	"fmt"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("facebook", New())
}

type extractor struct{}

// New returns a facebook extractor.
func New() extractors.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	var err error
	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, err
	}
	titles := utils.MatchOneOf(html, `<title id="pageTitle">(.+)</title>`)
	if titles == nil || len(titles) < 2 {
		return nil, extractors.ErrURLParseFailed
	}
	title := titles[1]

	streams := make(map[string]*extractors.Stream, 2)
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
		urlData := &extractors.Part{
			URL:  u,
			Size: size,
			Ext:  "mp4",
		}
		streams[quality] = &extractors.Stream{
			Parts:   []*extractors.Part{urlData},
			Size:    size,
			Quality: quality,
		}
	}

	return []*extractors.Data{
		{
			Site:    "Facebook facebook.com",
			Title:   title,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
