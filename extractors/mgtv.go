package extractors

import (
	"encoding/json"
	// "fmt"
	"strconv"
	"strings"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type mgtvVideoStream struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Def  string `json:"def"`
}

type mgtvVideoInfo struct {
	Title string `json:"title"`
	Desc  string `json:"desc"`
}

type mgtvVideoData struct {
	Stream       []mgtvVideoStream `json:"stream"`
	StreamDomain []string          `json:"stream_domain"`
	Info         mgtvVideoInfo     `json:"info"`
}

type mgtv struct {
	Data mgtvVideoData `json:"data"`
}

type mgtvVideoAddr struct {
	Info string `json:"info"`
}

type mgtvURLInfo struct {
	URL  string
	Size int64
}

func mgtvM3u8(url string) ([]mgtvURLInfo, int64) {
	var data []mgtvURLInfo
	var temp mgtvURLInfo
	var size, totalSize int64
	urls := utils.M3u8URLs(url)
	m3u8String := request.Get(url)
	sizes := utils.MatchAll(m3u8String, `#EXT-MGTV-File-SIZE:(\d+)`)
	// sizes: [[#EXT-MGTV-File-SIZE:1893724, 1893724]]
	for index, u := range urls {
		size, _ = strconv.ParseInt(sizes[index][1], 10, 64)
		totalSize += size
		temp = mgtvURLInfo{
			URL:  u,
			Size: size,
		}
		data = append(data, temp)
	}
	return data, totalSize
}

// Mgtv download function
func Mgtv(url string) downloader.VideoData {
	html := request.Get(url)
	vid := utils.MatchOneOf(
		url,
		`https?://www.mgtv.com/(?:b|l)/\d+/(\d+).html`,
		`https?://www.mgtv.com/hz/bdpz/\d+/(\d+).html`,
	)
	if vid == nil {
		vid = utils.MatchOneOf(html, `vid: (\d+),`)
	}
	dataString := request.Get("https://pcweb.api.mgtv.com/player/video?video_id=" + vid[1])
	var mgtvData mgtv
	json.Unmarshal([]byte(dataString), &mgtvData)
	title := strings.TrimSpace(
		mgtvData.Data.Info.Title + " " + mgtvData.Data.Info.Desc,
	)
	streams := mgtvData.Data.Stream
	var addr mgtvVideoAddr
	format := map[string]downloader.FormatData{}
	for _, stream := range streams {
		// real download address
		addr = mgtvVideoAddr{}
		json.Unmarshal(
			[]byte(request.Get(mgtvData.Data.StreamDomain[0]+stream.URL)), &addr,
		)
		m3u8URLs, totalSize := mgtvM3u8(addr.Info)
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
		format[stream.Def] = downloader.FormatData{
			URLs:    urls,
			Size:    totalSize,
			Quality: stream.Name,
		}
	}
	// best quality
	for _, quality := range []string{"3", "2", "1"} {
		// 超清，高清，标清
		if data, ok := format[quality]; ok {
			format["default"] = data
			delete(format, quality)
			break
		}
	}
	extractedData := downloader.VideoData{
		Site:    "芒果TV mgtv.com",
		Title:   title,
		Type:    "video",
		Formats: format,
	}
	extractedData.Download(url)
	return extractedData
}
