package bilibili

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/parser"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

const (
	bilibiliAPI        = "https://interface.bilibili.com/v2/playurl?"
	bilibiliBangumiAPI = "https://bangumi.bilibili.com/player/web_api/v2/playurl?"
	bilibiliTokenAPI   = "https://api.bilibili.com/x/player/playurl/token?"
)

const (
	// BiliBili blocks keys from time to time.
	// You can extract from the Android client or bilibiliPlayer.min.js
	appKey = "84956560bc028eb7"
	secKey = "94aba54af9065f71de72f5508f1cd42e"
)

const referer = "https://www.bilibili.com"

var utoken string

func genAPI(aid, cid string, bangumi bool, quality string, seasonType string) (string, error) {
	var (
		err        error
		baseAPIURL string
		params     string
	)
	if config.Cookie != "" && utoken == "" {
		utoken, err = request.Get(
			fmt.Sprintf("%said=%s&cid=%s", bilibiliTokenAPI, aid, cid),
			referer,
			nil,
		)
		if err != nil {
			return "", err
		}
		var t token
		json.Unmarshal([]byte(utoken), &t)
		if t.Code != 0 {
			return "", fmt.Errorf("Cookie error: %s", t.Message)
		}
		utoken = t.Data.Token
	}
	if bangumi {
		// The parameters need to be sorted by name
		// qn=0 flag makes the CDN address different every time
		// quality=116(1080P 60) is the highest quality so far
		params = fmt.Sprintf(
			"appkey=%s&cid=%s&module=bangumi&otype=json&qn=%s&quality=%s&season_type=%s&type=",
			appKey, cid, quality, quality, seasonType,
		)
		baseAPIURL = bilibiliBangumiAPI
	} else {
		params = fmt.Sprintf(
			"appkey=%s&cid=%s&otype=json&qn=%s&quality=%s&type=",
			appKey, cid, quality, quality,
		)
		baseAPIURL = bilibiliAPI
	}
	// bangumi utoken also need to put in params to sign, but the ordinary video doesn't need
	api := fmt.Sprintf(
		"%s%s&sign=%s", baseAPIURL, params, utils.Md5(params+secKey),
	)
	if !bangumi && utoken != "" {
		api = fmt.Sprintf("%s&utoken=%s", api, utoken)
	}
	return api, nil
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
	P        int
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
func Download(url string) ([]downloader.VideoData, error) {
	var options bilibiliOptions
	var err error
	if strings.Contains(url, "bangumi") {
		options.Bangumi = true
	}
	html, err := request.Get(url, referer, nil)
	if err != nil {
		return downloader.EmptyData, err
	}
	if !config.Playlist {
		options.HTML = html
		pageData, err := getMultiPageData(html)
		if err == nil && !options.Bangumi {
			// handle URL that has a playlist, mainly for unified titles
			// <h1> tag does not include subtitles
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
			options.P = p
			page := pageData.VideoData.Pages[p-1]
			options.Aid = pageData.Aid
			options.Cid = strconv.Itoa(page.Cid)
			// "part":"" or "part":"Untitled"
			if page.Part == "Untitled" {
				options.Subtitle = ""
			} else {
				options.Subtitle = page.Part
			}
		}
		data, err := bilibiliDownload(url, options)
		if err != nil {
			return downloader.EmptyData, err
		}
		return []downloader.VideoData{data}, nil
	}
	// for Bangumi playlist
	if options.Bangumi {
		dataString := utils.MatchOneOf(html, `window.__INITIAL_STATE__=(.+?);\(function`)[1]
		var data bangumiData
		json.Unmarshal([]byte(dataString), &data)
		needDownloadItems := utils.NeedDownloadList(len(data.EpList))
		extractedData := make([]downloader.VideoData, len(needDownloadItems))
		for index, u := range data.EpList {
			if !utils.ItemInSlice(index+1, needDownloadItems) {
				continue
			}
			videoData, err := bilibiliDownload(
				fmt.Sprintf("https://www.bilibili.com/bangumi/play/ep%d", u.EpID), options,
			)
			if err == nil {
				// if err is not nil, the data is empty struct
				extractedData[index] = videoData
			}
		}
		return extractedData, nil
	}
	// for normal video playlist
	data, err := getMultiPageData(html)
	if err != nil {
		// this page has no playlist
		options.HTML = html
		videoData, err := bilibiliDownload(url, options)
		if err != nil {
			return downloader.EmptyData, err
		}
		return []downloader.VideoData{videoData}, nil
	}
	// https://www.bilibili.com/video/av20827366/?p=1
	needDownloadItems := utils.NeedDownloadList(len(data.VideoData.Pages))
	extractedData := make([]downloader.VideoData, len(needDownloadItems))
	for index, u := range data.VideoData.Pages {
		if !utils.ItemInSlice(index+1, needDownloadItems) {
			continue
		}
		options.Aid = data.Aid
		options.Cid = strconv.Itoa(u.Cid)
		options.Subtitle = u.Part
		options.P = u.Page
		videoData, err := bilibiliDownload(url, options)
		if err == nil {
			extractedData[index] = videoData
		}
	}
	return extractedData, nil
}

func bilibiliDownload(url string, options bilibiliOptions) (downloader.VideoData, error) {
	var (
		aid, cid, html string
		err            error
	)
	if options.HTML != "" {
		// reuse html string, but this can't be reused in case of playlist
		html = options.HTML
	} else {
		html, err = request.Get(url, referer, nil)
		if err != nil {
			return downloader.VideoData{}, err
		}
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
	var seasonType string
	if options.Bangumi {
		seasonType = utils.MatchOneOf(html, `"season_type":(\d+)`)[1]
	}

	// Get "accept_quality" and "accept_description"
	// "accept_description":["高清 1080P","高清 720P","清晰 480P","流畅 360P"],
	// "accept_quality":[80,48,32,16],
	api, err := genAPI(aid, cid, options.Bangumi, "15", seasonType)
	if err != nil {
		return downloader.VideoData{}, err
	}
	jsonString, err := request.Get(api, referer, nil)
	if err != nil {
		return downloader.VideoData{}, err
	}
	var quality qualityInfo
	json.Unmarshal([]byte(jsonString), &quality)

	format := make(map[string]downloader.FormatData, len(quality.Quality))
	for _, q := range quality.Quality {
		apiURL, err := genAPI(aid, cid, options.Bangumi, strconv.Itoa(q), seasonType)
		if err != nil {
			return downloader.VideoData{}, err
		}
		jsonString, err := request.Get(apiURL, referer, nil)
		if err != nil {
			return downloader.VideoData{}, err
		}
		var data bilibiliData
		json.Unmarshal([]byte(jsonString), &data)

		// Avoid duplicate formats
		if _, ok := format[strconv.Itoa(data.Quality)]; ok {
			continue
		}

		urls, size := genURL(data.DURL)
		format[strconv.Itoa(data.Quality)] = downloader.FormatData{
			URLs:    urls,
			Size:    size,
			Quality: qualityString[data.Quality],
		}
	}

	// get the title
	doc, err := parser.GetDoc(html)
	if err != nil {
		return downloader.VideoData{}, err
	}
	title := parser.Title(doc)
	if options.Subtitle != "" {
		tempTitle := fmt.Sprintf("%s %s", title, options.Subtitle)
		if len([]rune(tempTitle)) > utils.MAXLENGTH {
			tempTitle = fmt.Sprintf("%s P%d %s", title, options.P, options.Subtitle)
		}
		title = tempTitle
	}
	title = utils.FileName(title)

	downloader.Caption(
		fmt.Sprintf("https://comment.bilibili.com/%s.xml", cid),
		url, title, "xml",
	)

	return downloader.VideoData{
		Site:    "哔哩哔哩 bilibili.com",
		Title:   title,
		Type:    "video",
		Formats: format,
	}, nil
}
