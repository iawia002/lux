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
			return "", fmt.Errorf("cookie error: %s", t.Message)
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

func genURL(durl []dURLData) ([]downloader.URL, int64) {
	var size int64
	urls := make([]downloader.URL, len(durl))
	for index, data := range durl {
		size += data.Size
		urls[index] = downloader.URL{
			URL:  data.URL,
			Size: data.Size,
			Ext:  "flv",
		}
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
		return data, errors.New("this page has no playlist")
	}
	json.Unmarshal([]byte(multiPageDataString[1]), &data)
	return data, nil
}

// Download bilibili main download function
func Download(url string) ([]downloader.Data, error) {
	var options bilibiliOptions
	var err error
	if strings.Contains(url, "bangumi") {
		options.Bangumi = true
	}
	html, err := request.Get(url, referer, nil)
	if err != nil {
		return downloader.EmptyList, err
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
		return []downloader.Data{bilibiliDownload(url, options)}, nil
	}
	// for Bangumi playlist
	if options.Bangumi {
		dataString := utils.MatchOneOf(html, `window.__INITIAL_STATE__=(.+?);\(function`)[1]
		var data bangumiData
		json.Unmarshal([]byte(dataString), &data)
		needDownloadItems := utils.NeedDownloadList(len(data.EpList))
		extractedData := make([]downloader.Data, len(needDownloadItems))
		wgp := utils.NewWaitGroupPool(config.ThreadNumber)
		dataIndex := 0
		for index, u := range data.EpList {
			if !utils.ItemInSlice(index+1, needDownloadItems) {
				continue
			}
			wgp.Add()
			go func(index, epID int, options bilibiliOptions, extractedData []downloader.Data) {
				defer wgp.Done()
				extractedData[index] = bilibiliDownload(
					fmt.Sprintf("https://www.bilibili.com/bangumi/play/ep%d", epID), options,
				)
			}(dataIndex, u.EpID, options, extractedData)
			dataIndex++
		}
		wgp.Wait()
		return extractedData, nil
	}
	// for normal video playlist
	data, err := getMultiPageData(html)
	if err != nil {
		// this page has no playlist
		options.HTML = html
		return []downloader.Data{bilibiliDownload(url, options)}, nil
	}
	// https://www.bilibili.com/video/av20827366/?p=1
	needDownloadItems := utils.NeedDownloadList(len(data.VideoData.Pages))
	extractedData := make([]downloader.Data, len(needDownloadItems))
	wgp := utils.NewWaitGroupPool(config.ThreadNumber)
	dataIndex := 0
	for index, u := range data.VideoData.Pages {
		if !utils.ItemInSlice(index+1, needDownloadItems) {
			continue
		}
		options.Aid = data.Aid
		options.Cid = strconv.Itoa(u.Cid)
		options.Subtitle = u.Part
		options.P = u.Page
		wgp.Add()
		go func(index int, url string, options bilibiliOptions, extractedData []downloader.Data) {
			defer wgp.Done()
			extractedData[index] = bilibiliDownload(url, options)
		}(dataIndex, url, options, extractedData)
		dataIndex++
	}
	wgp.Wait()
	return extractedData, nil
}

// bilibiliDownload download function for single url
func bilibiliDownload(url string, options bilibiliOptions) downloader.Data {
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
			return downloader.EmptyData(url, err)
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
			cid = utils.MatchOneOf(html, `cid=(\d+)`, `"cid":(\d+)`)[1]
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
		return downloader.EmptyData(url, err)
	}
	jsonString, err := request.Get(api, referer, nil)
	if err != nil {
		return downloader.EmptyData(url, err)
	}
	var quality qualityInfo
	json.Unmarshal([]byte(jsonString), &quality)

	streams := make(map[string]downloader.Stream, len(quality.Quality))
	for _, q := range quality.Quality {
		apiURL, err := genAPI(aid, cid, options.Bangumi, strconv.Itoa(q), seasonType)
		if err != nil {
			return downloader.EmptyData(url, err)
		}
		jsonString, err := request.Get(apiURL, referer, nil)
		if err != nil {
			return downloader.EmptyData(url, err)
		}
		var data bilibiliData
		json.Unmarshal([]byte(jsonString), &data)

		// Avoid duplicate streams
		if _, ok := streams[strconv.Itoa(data.Quality)]; ok {
			continue
		}

		urls, size := genURL(data.DURL)
		streams[strconv.Itoa(data.Quality)] = downloader.Stream{
			URLs:    urls,
			Size:    size,
			Quality: qualityString[data.Quality],
		}
	}

	// get the title
	doc, err := parser.GetDoc(html)
	if err != nil {
		return downloader.EmptyData(url, err)
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

	return downloader.Data{
		Site:    "哔哩哔哩 bilibili.com",
		Title:   title,
		Type:    "video",
		Streams: streams,
		URL:     url,
	}
}
