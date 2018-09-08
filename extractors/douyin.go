package extractors

import (
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type douyinVideoURLData struct {
	URLList []string `json:"url_list"`
}

type douyinVideoData struct {
	PlayAddr douyinVideoURLData `json:"play_addr"`
}

type douyinData struct {
	Video douyinVideoData `json:"video"`
	Desc  string          `json:"desc"`
}

// Douyin download function
func Douyin(url string) (downloader.VideoData, error) {
	var err error
	html, err := request.Get(url, url, nil)
	if err != nil {
		return downloader.VideoData{}, err
	}
	title := utils.MatchOneOf(html, `<p class="desc">(.+?)</p>`)[1]
	realURL := utils.MatchOneOf(html, `playAddr: "(.+?)"`)[1]
	size, err := request.Size(realURL, url)
	if err != nil {
		return downloader.VideoData{}, err
	}
	urlData := downloader.URLData{
		URL:  realURL,
		Size: size,
		Ext:  "mp4",
	}
	format := map[string]downloader.FormatData{
		"default": {
			URLs: []downloader.URLData{urlData},
			Size: size,
		},
	}
	extractedData := downloader.VideoData{
		Site:    "抖音 douyin.com",
		Title:   utils.FileName(title),
		Type:    "video",
		Formats: format,
	}
	err = extractedData.Download(url)
	if err != nil {
		return downloader.VideoData{}, err
	}
	return extractedData, nil
}
