package udn

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("udn", New())
}

const (
	startFlag = `',
            mp4: '//`
	endFlag = `'
        },
        subtitles`
)

func getCDNUrl(html string) string {
	if cdnURLs := utils.MatchOneOf(html, startFlag+"(.+?)"+endFlag); len(cdnURLs) > 1 && cdnURLs[1] != "" {
		return cdnURLs[1]
	}
	return ""
}

func prepareEmbedURL(url string) string {
	if !strings.Contains(url, "https://video.udn.com/embed/") {
		newIDs := strings.Split(url, "/")
		if len(newIDs) < 1 {
			return ""
		}
		return "https://video.udn.com/embed/news/" + newIDs[len(newIDs)-1]
	}
	return url
}

type extractor struct{}

// New returns a udn extractor.
func New() extractors.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	url = prepareEmbedURL(url)
	if len(url) == 0 {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}

	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var title string
	desc := utils.MatchOneOf(html, `title: '(.+?)',
        link:`)
	if len(desc) > 1 {
		title = desc[1]
	} else {
		title = "udn"
	}
	cdnURL := getCDNUrl(html)
	if cdnURL == "" {
		return nil, errors.New("empty list")
	}
	srcURL, err := request.Get("http://"+cdnURL, url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	size, err := request.Size(srcURL, url)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	urlData := &extractors.Part{
		URL:  srcURL,
		Size: size,
		Ext:  "mp4",
	}
	quality := "normal"
	streams := map[string]*extractors.Stream{
		quality: {
			Parts:   []*extractors.Part{urlData},
			Size:    size,
			Quality: quality,
		},
	}
	return []*extractors.Data{
		{
			Site:    "udn udn.com",
			Title:   title,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
