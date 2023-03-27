package zhihu

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

const (
	videoURL = "www.zhihu.com/zvideo"
	api      = "https://lens.zhihu.com/api/v4/videos/"
)

func init() {
	extractors.Register("zhihu", New())
}

type extractor struct{}

func New() extractors.Extractor {
	return &extractor{}
}

func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	if !strings.Contains(url, videoURL) {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}

	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	videoID := utils.MatchOneOf(html, `"videoId":"(\d+)"`)
	titleMatch := utils.MatchOneOf(html, `<title.*?>(.*?)</title>`)

	if len(videoID) <= 1 {
		return nil, errors.New("zhihu video id extract failed")
	}

	title := "Unknown"
	if len(titleMatch) > 1 {
		title = titleMatch[1]
	}

	resp, err := request.GetByte(fmt.Sprintf("%s%s", api, videoID[1]), url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var data video
	if err = json.Unmarshal(resp, &data); err != nil {
		return nil, errors.WithStack(err)
	}

	streams := make(map[string]*extractors.Stream)
	resolutions := map[string]resolution{
		"FHD": data.PlayList.FHD,
		"HD":  data.PlayList.HD,
		"SD":  data.PlayList.SD,
	}

	for k, v := range resolutions {
		stream := &extractors.Stream{
			Parts: []*extractors.Part{
				{
					URL:  v.PlayURL,
					Size: v.Size,
					Ext:  v.Format,
				},
			},
			Size: v.Size,
		}
		streams[k] = stream
	}

	return []*extractors.Data{
		{
			Site:    "知乎 zhihu.com",
			Title:   title,
			Streams: streams,
			Type:    extractors.DataTypeVideo,
			URL:     url,
		},
	}, nil
}
