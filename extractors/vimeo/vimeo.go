package vimeo

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/iawia002/annie/downloader"
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

// Download main download function
func Download(url string) ([]downloader.Data, error) {
	var (
		html, vid string
		err       error
	)
	if strings.Contains(url, "player.vimeo.com") {
		html, err = request.Get(url, url, nil)
		if err != nil {
			return downloader.EmptyList, err
		}
	} else {
		vid = utils.MatchOneOf(url, `vimeo\.com/(\d+)`)[1]
		html, err = request.Get("https://player.vimeo.com/video/"+vid, url, nil)
		if err != nil {
			return downloader.EmptyList, err
		}
	}
	jsonString := utils.MatchOneOf(html, `var \w+\s?=\s?({.+?});`)[1]
	var vimeoData vimeo
	json.Unmarshal([]byte(jsonString), &vimeoData)
	streams := map[string]downloader.Stream{}
	var size int64
	var urlData downloader.URL
	for _, video := range vimeoData.Request.Files.Progressive {
		size, err = request.Size(video.URL, url)
		if err != nil {
			return downloader.EmptyList, err
		}
		urlData = downloader.URL{
			URL:  video.URL,
			Size: size,
			Ext:  "mp4",
		}
		streams[strconv.Itoa(video.Profile)] = downloader.Stream{
			URLs:    []downloader.URL{urlData},
			Size:    size,
			Quality: video.Quality,
		}
	}

	return []downloader.Data{
		{
			Site:    "Vimeo vimeo.com",
			Title:   vimeoData.Video.Title,
			Type:    "video",
			Streams: streams,
			URL:     url,
		},
	}, nil
}
