package extractors

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
	format := map[string]downloader.FormatData{}
	var size int64
	var urlData downloader.URLData
	var highestProfile int
	for _, video := range vimeoData.Request.Files.Progressive {
		if video.Profile > highestProfile {
			highestProfile = video.Profile
		}
		size = request.Size(video.URL, url)
		urlData = downloader.URLData{
			URL:  video.URL,
			Size: size,
			Ext:  "mp4",
		}
		format[strconv.Itoa(video.Profile)] = downloader.FormatData{
			URLs:    []downloader.URLData{urlData},
			Size:    size,
			Quality: video.Quality,
		}
	}
	format["default"] = format[strconv.Itoa(highestProfile)]
	delete(format, strconv.Itoa(highestProfile))

	extractedData := downloader.VideoData{
		Site:    "Vimeo vimeo.com",
		Title:   vimeoData.Video.Title,
		Type:    "video",
		Formats: format,
	}
	extractedData.Download(url)
	return extractedData
}
