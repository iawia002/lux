package extractors

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/parser"
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

type data struct {
	Error  errorData `json:"error"`
	Stream []stream  `json:"stream"`
}

type youkuData struct {
	Data data `json:"data"`
}

var ccodes = []string{"0507", "0508", "0512", "0513", "0514", "0503", "0502", "0590"}
var referer = "https://v.youku.com"

func youkuUps(vid string) youkuData {
	var url string
	var utid string
	var html string
	var data youkuData
	headers := request.Headers("http://log.mmstat.com/eg.js", referer)
	setCookie := headers.Get("Set-Cookie")
	utid = utils.MatchOneOf(setCookie, `cna=(.+?);`)[1]
	for _, ccode := range ccodes {
		url = fmt.Sprintf(
			"https://ups.youku.com/ups/get.json?vid=%s&ccode=%s&client_ip=192.168.1.1&client_ts=%d&utid=%s",
			vid, ccode, time.Now().Unix(), utid,
		)
		html = request.Get(url)
		// data must be emptied before reassignment, otherwise it will contain the previous value(the 'error' data)
		data = youkuData{}
		json.Unmarshal([]byte(html), &data)
		if data.Data.Error.Code != -6004 {
			return data
		}
	}
	return data
}

func genData(youkuData data) ([]downloader.URLData, int64, string) {
	var (
		urls  []downloader.URLData
		size  int64
		index int
	)
	// get the best quality
	for i, s := range youkuData.Stream {
		if s.Size > size {
			size = s.Size
			index = i
		}
	}
	stream := youkuData.Stream[index]
	ext := strings.Split(
		strings.Split(stream.Segs[0].URL, "?")[0],
		".",
	)
	for _, data := range stream.Segs {
		url := downloader.URLData{
			URL:  data.URL,
			Size: data.Size,
			Ext:  ext[len(ext)-1],
		}
		urls = append(urls, url)
	}
	quality := fmt.Sprintf("%s %dx%d", stream.Type, stream.Width, stream.Height)
	return urls, stream.Size, quality
}

// Youku download function
func Youku(url string) downloader.VideoData {
	html := request.Get(url)
	// get the title
	doc := parser.GetDoc(html)
	title := parser.Title(doc)
	vid := utils.MatchOneOf(url, `id_(.+?).html`)[1]
	youkuData := youkuUps(vid)
	if youkuData.Data.Error.Code != 0 {
		log.Fatal(youkuData.Data.Error.Note)
	}
	urls, size, quality := genData(youkuData.Data)
	data := downloader.VideoData{
		Site:    "优酷 youku.com",
		Title:   title,
		Type:    "video",
		URLs:    urls,
		Size:    size,
		Quality: quality,
	}
	data.Download(url)
	return data
}
