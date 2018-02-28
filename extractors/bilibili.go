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

type dURLData struct {
	Size  int64  `json:"size"`
	URL   string `json:"url"`
	Order int    `json:"order"`
}

type bilibiliData struct {
	DURL    []dURLData `json:"durl"`
	Format  string     `json:"format"`
	Quality int        `json:"quality"`
}

func getSign(params string) string {
	sign := md5.New()
	sign.Write([]byte(params + secKey))
	return fmt.Sprintf("%x", sign.Sum(nil))
}

func genAPI(cid string, bangumi bool) string {
	var (
		baseAPIURL string
		params     string
	)
	if bangumi {
		// The parameters need to be sorted by name
		// qn=0 flag makes the CDN address different every time
		params = fmt.Sprintf(
			"appkey=%s&cid=%s&module=bangumi&otype=json&quality=0&season_type=4&type=",
			appKey, cid,
		)
		baseAPIURL = config.BILIBILI_BANGUMI_API
	} else {
		params = fmt.Sprintf(
			"appkey=%s&cid=%s&otype=json&quality=0&type=",
			appKey, cid,
		)
		baseAPIURL = config.BILIBILI_API
	}
	api := baseAPIURL + params + "&sign=" + getSign(params)
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
		}
		urls = append(urls, url)
	}
	return urls, size
}

// Bilibili download function
func Bilibili(url string) downloader.VideoData {
	var (
		bangumi bool
		cid     string
	)
	if strings.Contains(url, "bangumi") {
		bangumi = true
	}
	html := request.Get(url)
	if bangumi {
		cid = utils.Match1(`"cid":(\d+)`, html)[1]
	} else {
		cid = utils.Match1(`cid=(\d+)`, html)[1]
	}
	api := genAPI(cid, bangumi)
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
	format := dataDict.Format
	if format == "flv720" {
		format = "flv"
	}
	data := downloader.VideoData{
		Site:  "哔哩哔哩 bilibili.com",
		Title: title,
		URLs:  urls,
		Ext:   format,
		Size:  size,
	}
	data.Download(url)
	return data
}
