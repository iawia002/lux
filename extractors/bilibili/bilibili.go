package bilibili

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/parser"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

const (
	// BiliBili blocks keys from time to time.
	// You can extract from the Android client or bilibiliPlayer.min.js
	appKey string = "84956560bc028eb7"
	secKey string = "94aba54af9065f71de72f5508f1cd42e"
)

func genAPI(aid, cid string, bangumi bool) string {
	var (
		baseAPIURL string
		params     string
	)
	utoken := ""
	if config.Cookie != "" {
		utoken = request.Get(fmt.Sprintf(
			"%said=%s&cid=%s", config.BILIBILI_TOKEN_API, aid, cid,
		))
		var t token
		json.Unmarshal([]byte(utoken), &t)
		if t.Code != 0 {
			log.Println(config.Cookie)
			log.Fatal("Cookie error: ", t.Message)
		}
		utoken = t.Data.Token
	}
	if bangumi {
		// The parameters need to be sorted by name
		// qn=0 flag makes the CDN address different every time
		// quality=116(1080P 60) is the highest quality so far
		params = fmt.Sprintf(
			"appkey=%s&cid=%s&module=bangumi&otype=json&qn=116&quality=116&season_type=4&type=&utoken=%s",
			appKey, cid, utoken,
		)
		baseAPIURL = config.BILIBILI_BANGUMI_API
	} else {
		params = fmt.Sprintf(
			"appkey=%s&cid=%s&otype=json&qn=116&quality=116&type=",
			appKey, cid,
		)
		baseAPIURL = config.BILIBILI_API
	}
	// bangumi utoken also need to put in params to sign, but the ordinary video doesn't need
	api := fmt.Sprintf(
		"%s%s&sign=%s", baseAPIURL, params, utils.Md5(params+secKey),
	)
	if !bangumi && utoken != "" {
		api = fmt.Sprintf("%s&utoken=%s", api, utoken)
	}
	return api
}

func genURL(durl []dURLData) ([]downloader.URLData, int64) {
	var (
		urls []downloader.URLData
		size int64
	)
	for _, data := range durl {
		size += data.Size
		url := downloader.URLData{
			URL:  data.URL,
			Size: data.Size,
			Ext:  "flv",
		}
		urls = append(urls, url)
	}
	return urls, size
}

type bilibiliOptions struct {
	Bangumi  bool
	Subtitle string
	Aid      string
	Cid      string
	HTML     string
}

func getMultiPageData(html string) (multiPage, error) {
	var data multiPage
	multiPageDataString := utils.MatchOneOf(
		html, `window.__INITIAL_STATE__=(.+?);\(function`,
	)
	if multiPageDataString == nil {
		return data, errors.New("This page has no playlist")
	}
	json.Unmarshal([]byte(multiPageDataString[1]), &data)
	return data, nil
}

// Download bilibili main download function
func Download(url string) {
	var options bilibiliOptions
	if strings.Contains(url, "bangumi") {
		options.Bangumi = true
	}
	html := request.Get(url)
	if !config.Playlist {
		options.HTML = html
		data, err := getMultiPageData(html)
		if err == nil && !options.Bangumi {
			// handle URL that has a playlist, mainly for unified titles
			// bangumi doesn't need this
			pageString := utils.MatchOneOf(url, `\?p=(\d+)`)
			var p int
			if pageString == nil {
				// https://www.bilibili.com/video/av20827366/
				p = 1
			} else {
				// https://www.bilibili.com/video/av20827366/?p=2
				p, _ = strconv.Atoi(pageString[1])
			}
			page := data.VideoData.Pages[p-1]
			options.Aid = data.Aid
			options.Cid = strconv.Itoa(page.Cid)
			options.Subtitle = page.Part
		}
		bilibiliDownload(url, options)
		return
	}
	if options.Bangumi {
		dataString := utils.MatchOneOf(html, `window.__INITIAL_STATE__=(.+?);`)[1]
		var data bangumiData
		json.Unmarshal([]byte(dataString), &data)
		for _, u := range data.EpList {
			bilibiliDownload(
				fmt.Sprintf("https://www.bilibili.com/bangumi/play/ep%d", u.EpID), options,
			)
		}
	} else {
		data, err := getMultiPageData(html)
		if err != nil {
			// this page has no playlist
			options.HTML = html
			bilibiliDownload(url, options)
			return
		}
		// https://www.bilibili.com/video/av20827366/?p=1
		for _, u := range data.VideoData.Pages {
			options.Aid = data.Aid
			options.Cid = strconv.Itoa(u.Cid)
			options.Subtitle = u.Part
			bilibiliDownload(url, options)
		}
	}
}

func bilibiliDownload(url string, options bilibiliOptions) downloader.VideoData {
	var (
		aid, cid, html string
	)
	if options.HTML != "" {
		// reuse html string, but this can't be reused in case of playlist
		html = options.HTML
	} else {
		html = request.Get(url)
	}
	if options.Aid != "" && options.Cid != "" {
		aid = options.Aid
		cid = options.Cid
	} else {
		if options.Bangumi {
			cid = utils.MatchOneOf(html, `"cid":(\d+)`)[1]
			aid = utils.MatchOneOf(html, `"aid":(\d+)`)[1]
		} else {
			cid = utils.MatchOneOf(html, `cid=(\d+)`)[1]
			aid = utils.MatchOneOf(url, `av(\d+)`)[1]
		}
	}
	api := genAPI(aid, cid, options.Bangumi)
	apiData := request.Get(api)
	var dataDict bilibiliData
	json.Unmarshal([]byte(apiData), &dataDict)

	// get the title
	doc := parser.GetDoc(html)
	title := parser.Title(doc)
	if options.Subtitle != "" {
		title = fmt.Sprintf("%s %s", title, options.Subtitle)
	}

	urls, size := genURL(dataDict.DURL)
	format := map[string]downloader.FormatData{
		"default": downloader.FormatData{
			URLs:    urls,
			Size:    size,
			Quality: quality[dataDict.Quality],
		},
	}
	extractedData := downloader.VideoData{
		Site:    "哔哩哔哩 bilibili.com",
		Title:   title,
		Type:    "video",
		Formats: format,
	}
	extractedData.Download(url)
	return extractedData
}
