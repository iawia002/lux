package extractors

import (
	"encoding/json"
	"fmt"
	"log"
	netURL "net/url"
	"strings"
	"time"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type errorData struct {
	Note string `json:"note"`
	Code int    `json:"code"`
}

type segs struct {
	Size int64  `json:"size"`
	URL  string `json:"cdn_url"`
}

type stream struct {
	Size   int64  `json:"size"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Segs   []segs `json:"segs"`
	Type   string `json:"stream_type"`
}

type youkuVideo struct {
	Title string `json:"title"`
}

type youkuShow struct {
	Title string `json:"title"`
}

type data struct {
	Error  errorData  `json:"error"`
	Stream []stream   `json:"stream"`
	Video  youkuVideo `json:"video"`
	Show   youkuShow  `json:"show"`
}

type youkuData struct {
	Data data `json:"data"`
}

const youkuReferer = "https://v.youku.com"

// var ccodes = []string{"0510", "0502", "0507", "0508", "0512", "0513", "0514", "0503", "0590"}

var ccodes = []string{"0510"}

func youkuUps(vid string) youkuData {
	var url string
	var utid string
	var html string
	var data youkuData
	headers := request.Headers("http://log.mmstat.com/eg.js", youkuReferer)
	setCookie := headers.Get("Set-Cookie")
	utid = utils.MatchOneOf(setCookie, `cna=(.+?);`)[1]
	// http://g.alicdn.com/player/ykplayer/0.5.28/youku-player.min.js
	// grep -oE '"[0-9a-zA-Z+/=]{256}"' youku-player.min.js
	ckey := "DIl58SLFxFNndSV1GFNnMQVYkx1PP5tKe1siZu/86PR1u/Wh1Ptd+WOZsHHWxysSfAOhNJpdVWsdVJNsfJ8Sxd8WKVvNfAS8aS8fAOzYARzPyPc3JvtnPHjTdKfESTdnuTW6ZPvk2pNDh4uFzotgdMEFkzQ5wZVXl2Pf1/Y6hLK0OnCNxBj3+nb0v72gZ6b0td+WOZsHHWxysSo/0y9D2K42SaB8Y/+aD2K42SaB8Y/+ahU+WOZsHcrxysooUeND"
	for _, ccode := range ccodes {
		url = fmt.Sprintf(
			"https://ups.youku.com/ups/get.json?vid=%s&ccode=%s&client_ip=192.168.1.1&client_ts=%d&utid=%s&ckey=%s",
			vid, ccode, time.Now().Unix()/1000, netURL.QueryEscape(utid), netURL.QueryEscape(ckey),
		)
		html = request.Get(url, youkuReferer)
		// data must be emptied before reassignment, otherwise it will contain the previous value(the 'error' data)
		data = youkuData{}
		json.Unmarshal([]byte(html), &data)
		if data.Data.Error == (errorData{}) {
			return data
		}
	}
	return data
}

func genData(youkuData data) map[string]downloader.FormatData {
	var (
		size        int64
		bestQuality string
	)
	format := map[string]downloader.FormatData{}
	// get the best quality
	for _, s := range youkuData.Stream {
		if s.Size > size {
			size = s.Size
			bestQuality = s.Type
		}
	}
	for _, stream := range youkuData.Stream {
		ext := strings.Split(
			strings.Split(stream.Segs[0].URL, "?")[0],
			".",
		)
		urls := []downloader.URLData{}
		for _, data := range stream.Segs {
			url := downloader.URLData{
				URL:  data.URL,
				Size: data.Size,
				Ext:  ext[len(ext)-1],
			}
			urls = append(urls, url)
		}
		quality := fmt.Sprintf("%s %dx%d", stream.Type, stream.Width, stream.Height)
		format[stream.Type] = downloader.FormatData{
			URLs:    urls,
			Size:    stream.Size,
			Quality: quality,
		}
	}
	format["default"] = format[bestQuality]
	delete(format, bestQuality)
	return format
}

// Youku download function
func Youku(url string) downloader.VideoData {
	vid := utils.MatchOneOf(url, `id_(.+?).html`)[1]
	youkuData := youkuUps(vid)
	if youkuData.Data.Error.Code != 0 {
		log.Fatal(youkuData.Data.Error.Note)
	}
	format := genData(youkuData.Data)
	var title string
	if youkuData.Data.Show.Title == "" {
		title = youkuData.Data.Video.Title
	} else {
		title = fmt.Sprintf("%s %s", youkuData.Data.Show.Title, youkuData.Data.Video.Title)
	}
	data := downloader.VideoData{
		Site:    "优酷 youku.com",
		Title:   utils.FileName(title),
		Type:    "video",
		Formats: format,
	}
	data.Download(url)
	return data
}
