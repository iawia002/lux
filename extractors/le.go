package extractors

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type leData struct {
	Data []struct {
		Infos []struct {
			Vwidth  int    `json:"vwidth"`
			Vheight int    `json:"vheight"`
			MainURL string `json:"mainUrl"`
			Gfmt    string `json:"gfmt"`
		} `json:"infos"`
	} `json:"data"`
	StatusCode int `json:"statusCode"`
}

var leQuality = map[string]string{
	"52": "1080P",
	"22": "超清",
	"13": "高清",
	"21": "标清",
}

// Le download function
func Le(url string) downloader.VideoData {
	vid := utils.MatchOneOf(url, `ptv/vplay/(\d+).html`)[1]
	var (
		apiURL      string
		jsonData    string
		bestQuality string
		data        leData
	)
	format := map[string]downloader.FormatData{}
	for _, q := range []string{"52", "22", "13", "21"} {
		apiURL = fmt.Sprintf(
			"http://tvepg.letv.com/apk/data/common/security/playurl/geturl/byvid.shtml?vid=%s&key=&vtype=%s",
			vid, q,
		)
		jsonData = request.Get(apiURL, url, nil)
		data = leData{}
		json.Unmarshal([]byte(jsonData), &data)
		if data.StatusCode != 1001 {
			log.Fatal("The video doesn't exist")
		}
		if data.Data == nil {
			continue
		}
		if bestQuality == "" {
			bestQuality = q
		}
		urls := []downloader.URLData{}
		var (
			totalSize int64
			size      int64
			urlData   downloader.URLData
		)
		info := data.Data[0].Infos[0]
		for _, ts := range utils.M3u8URLs(info.MainURL) {
			size, _ = strconv.ParseInt(
				utils.MatchOneOf(ts, `_(\d+)_\d+_\d+\.ts`)[1], 10, 64,
			)
			urlData = downloader.URLData{
				URL:  ts,
				Size: size,
				Ext:  "ts",
			}
			totalSize += size
			urls = append(urls, urlData)
		}
		format[q] = downloader.FormatData{
			URLs: urls,
			Size: totalSize,
			Quality: fmt.Sprintf(
				"%s %dx%d",
				leQuality[q], info.Vwidth, info.Vheight,
			),
		}
	}
	format["default"] = format[bestQuality]
	delete(format, bestQuality)
	html := request.Get(url, url, nil)
	title := utils.MatchOneOf(html, `title:"(.+?)",`)[1]
	extractedData := downloader.VideoData{
		Site:    "乐视 le.com",
		Title:   utils.FileName(title),
		Type:    "video",
		Formats: format,
	}
	extractedData.Download(url)
	return extractedData
}
