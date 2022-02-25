package facebook

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"

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
		return nil, errors.WithStack(err)
	}
	titles := utils.MatchOneOf(html, `<title>([^<]+)</title>`)
	if titles == nil || len(titles) < 2 {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}

	title := strings.TrimSpace(titles[1])

	title = regexp.MustCompile(`\n+`).ReplaceAllString(title, " ")

	qualityRegMap := map[string]*regexp.Regexp{
		"sd": regexp.MustCompile(`"playable_url":\s*"([^"]+)"`),
		// "hd": regexp.MustCompile(`"playable_url_quality_hd":\s*"([^"]+)"`),
	}

	streams := make(map[string]*extractors.Stream, 2)
	for quality, qualityReg := range qualityRegMap {
		matcher := qualityReg.FindStringSubmatch(html)

		if len(matcher) == 0 {
			continue
		}

		u := strings.ReplaceAll(matcher[1], "\\", "")

		size, err := request.Size(u, url)
		if err != nil {
			return nil, errors.WithStack(err)
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
