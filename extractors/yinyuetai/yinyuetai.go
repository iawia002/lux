package yinyuetai

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iawia002/annie/extractors/types"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

const yinyuetaiAPI = "https://ext.yinyuetai.com/main/"

const (
	actionGetMvInfo = "get-h-mv-info"
)

func genAPI(action string, param string) string {
	return fmt.Sprintf("%s%s?json=true&%s", yinyuetaiAPI, action, param)
}

type extractor struct{}

// New returns a youtube extractor.
func New() types.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option types.Options) ([]*types.Data, error) {
	vid := utils.MatchOneOf(
		url,
		`https?://v.yinyuetai.com/video/(\d+)(?:\?vid=\d+)?`,
		`https?://v.yinyuetai.com/video/h5/(\d+)(?:\?vid=\d+)?`,
		`https?://m2.yinyuetai.com/video.html\?id=(\d+)`,
	)
	if vid == nil || len(vid) < 2 {
		return nil, errors.New("invalid url for yinyuetai")
	}
	params := fmt.Sprintf("videoId=%s", vid[1])
	// generate api url
	apiURL := genAPI(actionGetMvInfo, params)
	var err error
	html, err := request.Get(apiURL, url, nil)
	if err != nil {
		return nil, err
	}
	// parse yinyuetai data
	data := yinyuetaiMvData{}
	if err = json.Unmarshal([]byte(html), &data); err != nil {
		return nil, types.ErrURLParseFailed
	}
	// handle api error
	if data.Error {
		return nil, errors.New(data.Message)
	}
	if data.VideoInfo.CoreVideoInfo.Error {
		return nil, errors.New(data.VideoInfo.CoreVideoInfo.ErrorMsg)
	}
	title := data.VideoInfo.CoreVideoInfo.VideoName
	streams := make(map[string]*types.Stream, len(data.VideoInfo.CoreVideoInfo.VideoURLModels))
	// set streams
	for _, model := range data.VideoInfo.CoreVideoInfo.VideoURLModels {
		urlData := &types.Part{
			URL:  model.VideoURL,
			Size: model.FileSize,
			Ext:  "mp4",
		}
		streams[model.QualityLevel] = &types.Stream{
			Parts:   []*types.Part{urlData},
			Size:    model.FileSize,
			Quality: model.QualityLevelName,
		}
	}
	return []*types.Data{
		{
			Site:    "音悦台 yinyuetai.com",
			Title:   title,
			Type:    types.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
