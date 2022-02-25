package hupu

import (
	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("hupu", New())
}

type extractor struct{}

// New returns a hupu extractor.
func New() extractors.Extractor {
	return &extractor{}
}

func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var title string
	titleDesc := utils.MatchOneOf(html, `<span class="post-user-comp-info-bottom-title">(.+?)</span>`)
	if len(titleDesc) > 1 {
		title = titleDesc[1]
	} else {
		title = "hupu video"
	}

	var videoUrl string
	urlDesc := utils.MatchOneOf(html, `<video src="(.+?)" controls="" poster=(.+?)></video>`)
	if len(urlDesc) > 1 {
		videoUrl = urlDesc[1]
	} else {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}

	size, err := request.Size(videoUrl, url)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	urlData := &extractors.Part{
		URL:  videoUrl,
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
			Site:    "虎扑 hupu.com",
			Title:   title,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
