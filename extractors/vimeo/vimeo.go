package vimeo

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("vimeo", New())
}

type vimeoProgressive struct {
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Profile string `json:"profile"`
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

// New returns a vimeo extractor.
func New() extractors.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	var (
		html, vid string
		err       error
	)
	if strings.Contains(url, "player.vimeo.com") {
		html, err = request.Get(url, url, nil)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	} else {
		vid = utils.MatchOneOf(url, `vimeo\.com/(\d+)`)[1]
		html, err = request.Get("https://player.vimeo.com/video/"+vid, url, nil)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	jsonStrings := utils.MatchOneOf(html, `var \w+\s?=\s?({.+?});`)
	if jsonStrings == nil || len(jsonStrings) < 2 {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}
	jsonString := jsonStrings[1]

	var vimeoData vimeo
	if err = json.Unmarshal([]byte(jsonString), &vimeoData); err != nil {
		return nil, errors.WithStack(err)
	}

	streams := make(map[string]*extractors.Stream, len(vimeoData.Request.Files.Progressive))
	var size int64
	for _, video := range vimeoData.Request.Files.Progressive {
		size, err = request.Size(video.URL, url)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		urlData := &extractors.Part{
			URL:  video.URL,
			Size: size,
			Ext:  "mp4",
		}
		streams[video.Profile] = &extractors.Stream{
			Parts:   []*extractors.Part{urlData},
			Size:    size,
			Quality: video.Quality,
		}
	}

	return []*extractors.Data{
		{
			Site:    "Vimeo vimeo.com",
			Title:   vimeoData.Video.Title,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
