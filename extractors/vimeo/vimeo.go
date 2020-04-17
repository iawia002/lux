package vimeo

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/iawia002/annie/extractors/types"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type vimeoProgressive struct {
	Profile int    `json:"profile"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Quality string `json:"quality"`
	URL     string `json:"url"`
}

type vimeoFiles struct {
	Progressive []vimeoProgressive `json:"progressive"`
}

type vimeoRequest struct {
	Files vimeoFiles `json:"files"`
}

type vimeoVideo struct {
	Title string `json:"title"`
}

type vimeo struct {
	Request vimeoRequest `json:"request"`
	Video   vimeoVideo   `json:"video"`
}

type extractor struct{}

// New returns a youtube extractor.
func New() types.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option types.Options) ([]*types.Data, error) {
	var (
		html, vid string
		err       error
	)
	if strings.Contains(url, "player.vimeo.com") {
		html, err = request.Get(url, url, nil)
		if err != nil {
			return nil, err
		}
	} else {
		vid = utils.MatchOneOf(url, `vimeo\.com/(\d+)`)[1]
		html, err = request.Get("https://player.vimeo.com/video/"+vid, url, nil)
		if err != nil {
			return nil, err
		}
	}
	jsonStrings := utils.MatchOneOf(html, `var \w+\s?=\s?({.+?});`)
	if jsonStrings == nil || len(jsonStrings) < 2 {
		return nil, types.ErrURLParseFailed
	}
	jsonString := jsonStrings[1]

	var vimeoData vimeo
	if err = json.Unmarshal([]byte(jsonString), &vimeoData); err != nil {
		return nil, err
	}

	streams := make(map[string]*types.Stream, len(vimeoData.Request.Files.Progressive))
	var size int64
	for _, video := range vimeoData.Request.Files.Progressive {
		size, err = request.Size(video.URL, url)
		if err != nil {
			return nil, err
		}
		urlData := &types.Part{
			URL:  video.URL,
			Size: size,
			Ext:  "mp4",
		}
		streams[strconv.Itoa(video.Profile)] = &types.Stream{
			Parts:   []*types.Part{urlData},
			Size:    size,
			Quality: video.Quality,
		}
	}

	return []*types.Data{
		{
			Site:    "Vimeo vimeo.com",
			Title:   vimeoData.Video.Title,
			Type:    types.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
