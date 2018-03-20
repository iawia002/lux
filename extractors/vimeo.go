package extractors

import (
	"encoding/json"
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

func bestQuality(progressives []vimeoProgressive) vimeoProgressive {
	var highestProfile int
	var data vimeoProgressive
	for _, progressive := range progressives {
		if progressive.Profile > highestProfile {
			highestProfile = progressive.Profile
			data = progressive
		}
	}
	return data
}

// Vimeo download function
func Vimeo(url string) downloader.VideoData {
	var html string
	var vid string
	if strings.Contains(url, "player.vimeo.com") {
		html = request.Get(url)
	} else {
		vid = utils.MatchOneOf(url, `vimeo\.com/(\d+)`)[1]
		html = request.Get("https://player.vimeo.com/video/" + vid)
	}
	jsonString := utils.MatchOneOf(html, `{var a=(.+?);`)[1]
	var vimeoData vimeo
	json.Unmarshal([]byte(jsonString), &vimeoData)
	video := bestQuality(vimeoData.Request.Files.Progressive)

	size := request.Size(video.URL, url)
	urlData := downloader.URLData{
		URL:  video.URL,
		Size: size,
		Ext:  "mp4",
	}
	data := downloader.VideoData{
		Site:    "Vimeo vimeo.com",
		Title:   vimeoData.Video.Title,
		Type:    "video",
		URLs:    []downloader.URLData{urlData},
		Size:    size,
		Quality: video.Quality,
	}
	data.Download(url)
	return data
}
