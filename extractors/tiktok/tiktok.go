package tiktok

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
)

func init() {
	extractors.Register("tiktok", New())
}

type extractor struct{}

// New returns a tiktok extractor.
func New() extractors.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	html, err := request.Get(url, url, map[string]string{
		// tiktok require a user agent
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:98.0) Gecko/20100101 Firefox/98.0",
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	urlMatcherRegExp := regexp.MustCompile(`"downloadAddr":\s*"([^"]+)"`)

	downloadURLMatcher := urlMatcherRegExp.FindStringSubmatch(html)

	if len(downloadURLMatcher) == 0 {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}

	videoURL := strings.ReplaceAll(downloadURLMatcher[1], `\u002F`, "/")

	titleMatcherRegExp := regexp.MustCompile(`<title[^>]*>([^<]+)</title>`)

	titleMatcher := titleMatcherRegExp.FindStringSubmatch(html)

	if len(titleMatcher) == 0 {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}

	title := titleMatcher[1]

	titleArr := strings.Split(title, "|")

	if len(titleArr) == 1 {
		title = titleArr[0]
	} else {
		title = strings.TrimSpace(strings.Join(titleArr[:len(titleArr)-1], "|"))
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
			Site:    "TikTok tiktok.com",
			Title:   title,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
