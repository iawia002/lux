package iqiyi

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/parser"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type vidl struct {
	M3utx      string `json:"m3utx"`
	Vd         int    `json:"vd"` // quality number
	ScreenSize string `json:"screenSize"`
}

type iqiyiData struct {
	Vidl []vidl `json:"vidl"`
}

type iqiyi struct {
	Code string    `json:"code"`
	Data iqiyiData `json:"data"`
}

const iqiyiReferer = "https://www.iqiyi.com"

func getIqiyiData(tvid, vid string) (iqiyi, error) {
	t := time.Now().Unix() * 1000
	src := "76f90cbd92f94a2e925d83e8ccd22cb7"
	key := "d5fb4bd9d50c4be6948c97edd7254b0e"
	sc := utils.Md5(strconv.FormatInt(t, 10) + key + vid)
	info, err := request.Get(
		fmt.Sprintf(
			"http://cache.m.iqiyi.com/jp/tmts/%s/%s/?t=%d&sc=%s&src=%s",
			tvid, vid, t, sc, src,
		),
		iqiyiReferer,
		nil,
	)
	if err != nil {
		return iqiyi{}, err
	}
	var data iqiyi
	json.Unmarshal([]byte(info[len("var tvInfoJs="):]), &data)
	return data, nil
}

// Download main download function
func Download(url string) ([]downloader.VideoData, error) {
	html, err := request.Get(url, iqiyiReferer, nil)
	if err != nil {
		return downloader.EmptyData, err
	}
	tvid := utils.MatchOneOf(
		url,
		`#curid=(.+)_`,
		`tvid=([^&]+)`,
	)
	if tvid == nil {
		tvid = utils.MatchOneOf(
			html,
			`data-player-tvid="([^"]+)"`,
			`param\['tvid'\]\s*=\s*"(.+?)"`,
			`"tvid":"(\d+)"`,
		)
	}
	vid := utils.MatchOneOf(
		url,
		`#curid=.+_(.*)$`,
		`vid=([^&]+)`,
	)
	if vid == nil {
		vid = utils.MatchOneOf(
			html,
			`data-player-videoid="([^"]+)"`,
			`param\['vid'\]\s*=\s*"(.+?)"`,
			`"vid":"(\w+)"`,
		)
	}
	doc, err := parser.GetDoc(html)
	if err != nil {
		return downloader.EmptyData, err
	}
	title := strings.TrimSpace(doc.Find("h1>a").First().Text())
	var sub string
	for _, k := range []string{"span", "em"} {
		if sub != "" {
			break
		}
		sub = strings.TrimSpace(doc.Find("h1>" + k).First().Text())
	}
	title += sub
	if title == "" {
		title = doc.Find("title").Text()
	}
	videoDatas, err := getIqiyiData(tvid[1], vid[1])
	if err != nil {
		return downloader.EmptyData, err
	}
	if videoDatas.Code != "A00000" {
		return downloader.EmptyData, errors.New("Can't play this video")
	}
	format := map[string]downloader.FormatData{}
	var urlData downloader.URLData
	var size, totalSize int64
	for _, video := range videoDatas.Data.Vidl {
		if video.Vd == 14 {
			// This format will go wrong when merging
			continue
		}
		urls := []downloader.URLData{}
		totalSize = 0
		m3u8URLs, err := utils.M3u8URLs(video.M3utx)
		if err != nil {
			return downloader.EmptyData, err
		}
		for _, ts := range m3u8URLs {
			size, _ = strconv.ParseInt(
				utils.MatchOneOf(ts, `contentlength=(\d+)`)[1], 10, 64,
			)
			// http://dx.data.video.qiyi.com -> http://data.video.qiyi.com
			urlData = downloader.URLData{
				URL:  strings.Replace(ts, "dx.data.video.qiyi.com", "data.video.qiyi.com", 1),
				Size: size,
				Ext:  "ts",
			}
			totalSize += size
			urls = append(urls, urlData)
		}
		format[strconv.Itoa(video.Vd)] = downloader.FormatData{
			URLs:    urls,
			Size:    totalSize,
			Quality: video.ScreenSize,
		}
	}

	return []downloader.VideoData{
		{
			Site:    "爱奇艺 iqiyi.com",
			Title:   utils.FileName(title),
			Type:    "video",
			Formats: format,
		},
	}, nil
}
