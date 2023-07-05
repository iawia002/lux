package pinterest

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
)

func init() {
	extractors.Register("pinterest", New())
}

type extractor struct{}

// New returns a pinterest extractor.
func New() extractors.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	html, err := request.Get(url, url, map[string]string{
		// pinterest require a user agent
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:98.0) Gecko/20100101 Firefox/98.0",
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	urlMatcherRegExp := regexp.MustCompile(`"contentUrl":"https:\/\/v1\.pinimg\.com\/videos\/mc\/720p\/[a-zA-Z0-9\/]+\.mp4`)

	downloadURLMatcher := urlMatcherRegExp.FindStringSubmatch(html)

	if len(downloadURLMatcher) == 0 {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}

	videoURL := strings.ReplaceAll(downloadURLMatcher[0], `"contentUrl":"`, "")

	titleMatcherRegExp := regexp.MustCompile(`<title[^>]*>([^<]+)</title>`)

	titleMatcher := titleMatcherRegExp.FindStringSubmatch(html)

	if len(titleMatcher) == 0 {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}

	title := strings.ReplaceAll(strings.ReplaceAll(titleMatcher[0], "<title>", ""), "</title>", "")

	titleArr := strings.Split(title, "|")

	if len(titleArr) > 0 {
		title = titleArr[0]
	}

	streams := make(map[string]*extractors.Stream)

	size, err := request.Size(videoURL, url)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	urlData := &extractors.Part{
		URL:  videoURL,
		Size: size,
		Ext:  "mp4",
	}
	streams["default"] = &extractors.Stream{
		Parts: []*extractors.Part{urlData},
		Size:  size,
	}

	return []*extractors.Data{
		{
			Site:    "Pinterest pinterest.com",
			Title:   title,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
