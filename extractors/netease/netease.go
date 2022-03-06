package netease

import (
	netURL "net/url"
	"strings"

	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("163", New())
}

type extractor struct{}

// New returns a netease extractor.
func New() extractors.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	url = strings.Replace(url, "/#/", "/", 1)
	vid := utils.MatchOneOf(url, `/(mv|video)\?id=(\w+)`)
	if vid == nil {
		return nil, errors.New("invalid url for netease music")
	}

	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if strings.Contains(html, "u-errlg-404") {
		return nil, errors.New("404 music not found")
	}

	titles := utils.MatchOneOf(html, `<meta property="og:title" content="(.+?)" />`)
	if titles == nil || len(titles) < 2 {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}
	title := titles[1]

	realURLs := utils.MatchOneOf(html, `<meta property="og:video" content="(.+?)" />`)
	if realURLs == nil || len(realURLs) < 2 {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}
	realURL, _ := netURL.QueryUnescape(realURLs[1])

	size, err := request.Size(realURL, url)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	urlData := &extractors.Part{
		URL:  realURL,
		Size: size,
		Ext:  "mp4",
	}
	streams := map[string]*extractors.Stream{
		"default": {
			Parts: []*extractors.Part{urlData},
			Size:  size,
		},
	}
	return []*extractors.Data{
		{
			Site:    "网易云音乐 music.163.com",
			Title:   title,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
