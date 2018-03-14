package extractors

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

const (
	// BiliBili blocks keys from time to time.
	// You can extract from the Android client or bilibiliPlayer.min.js
	appKey string = "84956560bc028eb7"
	secKey string = "94aba54af9065f71de72f5508f1cd42e"
)

func getSign(params string) string {
	sign := md5.New()
	sign.Write([]byte(params + secKey))
	return fmt.Sprintf("%x", sign.Sum(nil))
}

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
		"%s%s&sign=%s", baseAPIURL, params, getSign(params),
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

// Bilibili download function
func Bilibili(url string) {
	var bangumi bool
	if strings.Contains(url, "bangumi") {
		bangumi = true
	}
	if !config.Playlist {
		bilibiliDownload(url, bangumi)
		return
	}
	html := request.Get(url)
	if bangumi {
		dataString := utils.MatchOneOf(html, `window.__INITIAL_STATE__=(.+?);`)[1]
		var data bangumiData
		json.Unmarshal([]byte(dataString), &data)
		for _, u := range data.EpList {
			bilibiliDownload(
				fmt.Sprintf("https://www.bilibili.com/bangumi/play/ep%d", u.EpID), bangumi,
			)
		}
	} else {
		urls := utils.MatchAll(html, `<option value='(.+?)'`)
		if len(urls) == 0 {
			// this page has no playlist
			bilibiliDownload(url, bangumi)
			return
		}
		// /video/av16907446/index_1.html
		for _, u := range urls {
			bilibiliDownload("https://www.bilibili.com"+u[1], bangumi)
		}
	}
}

func bilibiliDownload(url string, bangumi bool) downloader.VideoData {
	var (
		aid, cid string
	)
	html := request.Get(url)
	if bangumi {
		cid = utils.MatchOneOf(html, `"cid":(\d+)`)[1]
		aid = utils.MatchOneOf(html, `"aid":(\d+)`)[1]
	} else {
		cid = utils.MatchOneOf(html, `cid=(\d+)`)[1]
		aid = utils.MatchOneOf(url, `av(\d+)`)[1]
	}
	api := genAPI(aid, cid, bangumi)
	apiData := request.Get(api)
	var dataDict bilibiliData
	json.Unmarshal([]byte(apiData), &dataDict)

	// get the title
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		log.Fatal(err)
	}
	var title string
	title = strings.TrimSpace(doc.Find("h1").First().Text())
	if title == "" {
		// Some movie page got no h1 tag
		title, _ = doc.Find("meta[property=\"og:title\"]").Attr("content")
	}

	urls, size := genURL(dataDict.DURL)
	data := downloader.VideoData{
		Site:    "哔哩哔哩 bilibili.com",
		Title:   utils.FileName(title),
		URLs:    urls,
		Type:    "video",
		Size:    size,
		Quality: quality[dataDict.Quality],
	}
	data.Download(url)
	return data
}
