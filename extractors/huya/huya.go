package huya

import (
	"github.com/iawia002/lux/extractors/types"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

type extractor struct{}

const huyaVideoHost = "https://videotx-platform.cdn.huya.com/"

// New returns a huya extractor.
func New() types.Extractor {
	return &extractor{}
}

func (e *extractor) Extract(url string, option types.Options) ([]*types.Data, error) {
	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, err
	}

	var title string
	titleDesc := utils.MatchOneOf(html, `<h1>(.+?)</h1>`)
	if len(titleDesc) > 1 {
		title = titleDesc[1]
	} else {
		title = "huya video"
	}

	var videoUrl string
	videoDesc := utils.MatchOneOf(html, `//videotx-platform.cdn.huya.com/(.*)" poster=(.+?)`)
	if len(videoDesc) > 1 {
		videoUrl = huyaVideoHost + videoDesc[1]
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
			Site:    "虎牙 huya.com",
			Title:   title,
			Type:    types.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
