package netease

import (
	"errors"
	netURL "net/url"
	"strings"

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
	url = strings.Replace(url, "/#/", "/", 1)
	vid := utils.MatchOneOf(url, `/(mv|video)\?id=(\w+)`)
	if vid == nil {
		return nil, errors.New("invalid url for netease music")
	}

	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, err
	}
	if strings.Contains(html, "u-errlg-404") {
		return nil, errors.New("404 music not found")
	}

	titles := utils.MatchOneOf(html, `<meta property="og:title" content="(.+?)" />`)
	if titles == nil || len(titles) < 2 {
		return nil, types.ErrURLParseFailed
	}
	title := titles[1]

	realURLs := utils.MatchOneOf(html, `<meta property="og:video" content="(.+?)" />`)
	if realURLs == nil || len(realURLs) < 2 {
		return nil, types.ErrURLParseFailed
	}
	realURL, _ := netURL.QueryUnescape(realURLs[1])

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
			Site:    "网易云音乐 music.163.com",
			Title:   title,
			Type:    types.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
