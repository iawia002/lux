package extractors

import (
	"encoding/json"
	"fmt"
	"log"
	netURL "net/url"
	"strings"
	"time"

	"github.com/iawia002/annie/config"
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
	Size      int64  `json:"size"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Segs      []segs `json:"segs"`
	Type      string `json:"stream_type"`
	AudioLang string `json:"audio_lang"`
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

func getAudioLang(lang string) string {
	var youkuAudioLang = map[string]string{
		"guoyu": "国语",
		"ja":    "日语",
	}
	translate, ok := youkuAudioLang[lang]
	if !ok {
		return lang
	}
	return translate
}

// https://g.alicdn.com/player/ykplayer/0.5.61/youku-player.min.js
// {"0505":"interior","050F":"interior","0501":"interior","0502":"interior","0503":"interior","0510":"adshow","0512":"BDskin","0590":"BDskin"}

// var ccodes = []string{"0510", "0502", "0507", "0508", "0512", "0513", "0514", "0503", "0590"}

func youkuUps(vid string) youkuData {
	var (
		url  string
		utid string
		html string
		data youkuData
	)
	if strings.Contains(config.Cookie, "cna=") {
		utid = utils.MatchOneOf(config.Cookie, `cna=(.+?);`)[1]
	} else {
		headers := request.Headers("http://log.mmstat.com/eg.js", youkuReferer)
		setCookie := headers.Get("Set-Cookie")
		utid = utils.MatchOneOf(setCookie, `cna=(.+?);`)[1]
	}
	// https://g.alicdn.com/player/ykplayer/0.5.61/youku-player.min.js
	// grep -oE '"[0-9a-zA-Z+/=]{256}"' youku-player.min.js
	ckey := "7B19C0AB12633B22E7FE81271162026020570708D6CC189E4924503C49D243A0DE6CD84A766832C2C99898FC5ED31F3709BB3CDD82C96492E721BDD381735026"
	for _, ccode := range []string{config.Ccode} {
		url = fmt.Sprintf(
			"https://ups.youku.com/ups/get.json?vid=%s&ccode=%s&client_ip=192.168.1.1&client_ts=%d&utid=%s&ckey=%s",
			vid, ccode, time.Now().Unix()/1000, netURL.QueryEscape(utid), netURL.QueryEscape(ckey),
		)
		html = request.Get(url, youkuReferer, nil)
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
		formatString string
		quality      string
	)
	format := map[string]downloader.FormatData{}
	for _, stream := range youkuData.Stream {
		if stream.AudioLang == "default" {
			formatString = stream.Type
			quality = fmt.Sprintf(
				"%s %dx%d", stream.Type, stream.Width, stream.Height,
			)
		} else {
			formatString = fmt.Sprintf("%s-%s", stream.Type, stream.AudioLang)
			quality = fmt.Sprintf(
				"%s %dx%d %s", stream.Type, stream.Width, stream.Height,
				getAudioLang(stream.AudioLang),
			)
		}

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
		format[formatString] = downloader.FormatData{
			URLs:    urls,
			Size:    stream.Size,
			Quality: quality,
		}
	}
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
	if youkuData.Data.Show.Title == "" || strings.Contains(
		youkuData.Data.Video.Title, youkuData.Data.Show.Title,
	) {
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
