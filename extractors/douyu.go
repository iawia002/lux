package extractors

import (
	"encoding/json"
	"log"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type douyuVideoData struct {
	VideoURL string `json:"video_url"`
}

type douyuData struct {
	Error int            `json:"error"`
	Data  douyuVideoData `json:"data"`
}

type douyuURLInfo struct {
	URL  string
	Size int64
}

func douyuM3u8(url string) ([]douyuURLInfo, int64) {
	var data []douyuURLInfo
	var temp douyuURLInfo
	var size, totalSize int64
	urls := utils.M3u8URLs(url)
	for _, u := range urls {
		size = request.Size(u, url)
		totalSize += size
		temp = douyuURLInfo{
			URL:  u,
			Size: size,
		}
		data = append(data, temp)
	}
	return data, totalSize
}

// Douyu download function
func Douyu(url string) downloader.VideoData {
	liveVid := utils.MatchOneOf(url, `https?://www.douyu.com/(\S+)`)
	if liveVid != nil {
		log.Fatal("暂不支持斗鱼直播")
	}

	html := request.Get(url)
	title := utils.MatchOneOf(html, `<title>(.*?)</title>`)[1]

	vid := utils.MatchOneOf(url, `https?://v.douyu.com/show/(\S+)`)[1]
	dataString := request.Get("http://vmobile.douyu.com/video/getInfo?vid=" + vid)
	var dataDict douyuData
	json.Unmarshal([]byte(dataString), &dataDict)

	m3u8URLs, totalSize := douyuM3u8(dataDict.Data.VideoURL)
	urls := []downloader.URLData{}
	var temp downloader.URLData
	for _, u := range m3u8URLs {
		temp = downloader.URLData{
			URL:  u.URL,
			Size: u.Size,
			Ext:  "ts",
		}
		urls = append(urls, temp)
	}

	format := map[string]downloader.FormatData{
		"default": downloader.FormatData{
			URLs: urls,
			Size: totalSize,
		},
	}
	extractedData := downloader.VideoData{
		Site:    "斗鱼 douyu.com",
		Title:   title,
		Type:    "video",
		Formats: format,
	}
	extractedData.Download(url)
	return extractedData
}
