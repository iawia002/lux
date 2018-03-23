package extractors

import (
	"encoding/json"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type douyinVideoURLData struct {
	URLList []string `json:"url_list"`
}

type douyinVideoData struct {
	PlayAddr     douyinVideoURLData `json:"play_addr"`
	RealPlayAddr string             `json:"real_play_addr"`
}

type douyinData struct {
	Video douyinVideoData `json:"video"`
	Desc  string          `json:"desc"`
}

// Douyin download function
func Douyin(url string) downloader.VideoData {
	html := request.Get(url)
	vData := utils.MatchOneOf(html, `var data = \[(.*?)\];`)[1]
	var dataDict douyinData
	json.Unmarshal([]byte(vData), &dataDict)

	size := request.Size(dataDict.Video.RealPlayAddr, url)
	urlData := downloader.URLData{
		URL:  dataDict.Video.RealPlayAddr,
		Size: size,
		Ext:  "mp4",
	}
	format := map[string]downloader.FormatData{
		"default": downloader.FormatData{
			URLs: []downloader.URLData{urlData},
			Size: size,
		},
	}
	extractedData := downloader.VideoData{
		Site:    "抖音 douyin.com",
		Title:   utils.FileName(dataDict.Desc),
		Type:    "video",
		Formats: format,
	}
	extractedData.Download(url)
	return extractedData
}
