package hupu

import (
	"github.com/iawia002/lux/extractors/types"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

type extractor struct{}

// New returns a hupu extractor.
func New() types.Extractor {
	return &extractor{}
}

func (e *extractor) Extract(url string, option types.Options) ([]*types.Data, error) {
	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, err
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
		return nil, types.ErrURLParseFailed
	}

	size, err := request.Size(videoUrl, url)
	if err != nil {
		return nil, err
	}
	urlData := &types.Part{
		URL:  videoUrl,
		Size: size,
		Ext:  "mp4",
	}
	quality := "normal"
	streams := map[string]*types.Stream{
		quality: {
			Parts:   []*types.Part{urlData},
			Size:    size,
			Quality: quality,
		},
	}
	return []*types.Data{
		{
			Site:    "虎扑 hupu.com",
			Title:   title,
			Type:    types.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
