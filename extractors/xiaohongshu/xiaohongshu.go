package xiaohongshu

import (
	"encoding/json"
	neturl "net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/iawia002/lux/config"
	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("xiaohongshu", New())
}

type extractor struct{}

// New returns a xiaohognshu extractor.
func New() extractors.Extractor {
	return &extractor{}
}

const mp4VideoType = "mp4"

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	html, err := request.Get(url, url, config.FakeHeaders)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// title
	titles := utils.MatchOneOf(html, `<title>(.*?)</title>`)
	if titles == nil || len(titles) != 2 {
		return nil, errors.WithStack(extractors.ErrBodyParseFailed)
	}
	title := titles[1]

	// video url
	urlsJSON := utils.MatchOneOf(html, `"backupUrls":(\[.+?\])`)
	if urlsJSON == nil || len(urlsJSON) != 2 {
		return nil, errors.WithStack(extractors.ErrBodyParseFailed)
	}
	var urls []string
	err = json.Unmarshal([]byte(urlsJSON[1]), &urls)
	if err != nil {
		return nil, errors.WithStack(extractors.ErrBodyParseFailed)
	}

	pUrl, err := neturl.ParseRequestURI(url)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// streams
	streams := make(map[string]*extractors.Stream)
	var size int64
	for i, u := range urls {
		if !strings.Contains(u, mp4VideoType) {
			continue
		}
		size, err = request.Size(u, u)
		if err != nil {
			continue
		}

		if pUrl.Host == "xhslink.com" && strings.Contains(u, "sns-video-qc") {
			size += 1 // Make sure the link is downloadable and sort the link first with the same size
		}
		streams[strconv.Itoa(i)] = &extractors.Stream{
			Parts: []*extractors.Part{
				{
					URL:  u,
					Size: size,
					Ext:  mp4VideoType,
				},
			},
			Size: size,
		}
	}
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if len(streams) == 0 {
		return nil, errors.WithStack(extractors.ErrBodyParseFailed)
	}

	return []*extractors.Data{
		{
			Site:    "小红书 xiaohongshu.com",
			Title:   title,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
