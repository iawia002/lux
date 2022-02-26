package tangdou

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("tangdou", New())
}

type extractor struct{}

// New returns a tangdou extractor.
func New() extractors.Extractor {
	return &extractor{}
}

var defaultHeader = map[string]string{
	"Sec-Fetch-Dest": "document",
	"Sec-Fetch-Mode": "navigate",
	"Sec-Fetch-Site": "cross-site",
	"Sec-GPC":        "1",
	"User-Agent":     "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:98.0) Gecko/20100101 Firefox/98.0",
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	return []*extractors.Data{tangdouDownload(url)}, nil
}

// tangdouDownload download function for single url
func tangdouDownload(uri string) *extractors.Data {
	html, err := request.Get(uri, uri, defaultHeader)
	if err != nil {
		return extractors.EmptyData(uri, err)
	}

	titles := utils.MatchOneOf(
		html, `<div class="title">(.+?)</div>`, `<meta name="description" content="(.+?)"`, `<title>(.+?)</title>`,
	)
	if titles == nil || len(titles) < 2 {
		return extractors.EmptyData(uri, errors.WithStack(extractors.ErrURLParseFailed))
	}
	title := titles[1]

	videoURLs := utils.MatchOneOf(
		html, `video:'(.+?)'`, `video:"(.+?)"`, `<video[^>]*src="(.+?)"`, `play_url:\s*"(.+?)",`,
	)

	if len(videoURLs) < 2 {
		return extractors.EmptyData(uri, errors.WithStack(extractors.ErrURLParseFailed))
	}

	realURL := strings.ReplaceAll(videoURLs[1], `\u002F`, "/")

	size, err := request.Size(realURL, uri)
	if err != nil {
		return extractors.EmptyData(uri, err)
	}

	streams := map[string]*extractors.Stream{
		"default": {
			Parts: []*extractors.Part{
				{
					URL:  realURL,
					Size: size,
					Ext:  "mp4",
				},
			},
			Size: size,
		},
	}

	return &extractors.Data{
		Site:    "糖豆广场舞 tangdou.com",
		Title:   title,
		Type:    extractors.DataTypeVideo,
		Streams: streams,
		URL:     uri,
	}
}
