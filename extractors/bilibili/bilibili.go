package bilibili

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/iawia002/annie/extractors/types"
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
	appKey = "iVGUTjsxvpLeuDCf"
	secKey = "aHRmhWMLkdeMuILqORnYZocwMBpMEOdt"
)

const referer = "https://www.bilibili.com"

var utoken string

func genAPI(aid, cid int, bangumi bool, quality, seasonType, cookie string) (string, error) {
	var (
		err        error
		baseAPIURL string
		params     string
	)
	if cookie != "" && utoken == "" {
		utoken, err = request.Get(
			fmt.Sprintf("%said=%d&cid=%d", bilibiliTokenAPI, aid, cid),
			referer,
			nil,
		)
		if err != nil {
			return "", err
		}
		var t token
		err = json.Unmarshal([]byte(utoken), &t)
		if err != nil {
			return "", err
		}
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
			"appkey=%s&cid=%d&module=bangumi&otype=json&qn=%s&quality=%s&season_type=%s&type=",
			appKey, cid, quality, quality, seasonType,
		)
		baseAPIURL = bilibiliBangumiAPI
	} else {
		params = fmt.Sprintf(
			"appkey=%s&cid=%d&otype=json&qn=%s&quality=%s&type=",
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

func genParts(durl []dURLData) ([]*types.Part, int64) {
	var size int64
	parts := make([]*types.Part, len(durl))
	for index, data := range durl {
		size += data.Size
		parts[index] = &types.Part{
			URL:  data.URL,
			Size: data.Size,
			Ext:  "flv",
		}
	}
	return parts, size
}

type bilibiliOptions struct {
	url      string
	html     string
	bangumi  bool
	aid      int
	cid      int
	page     int
	subtitle string
}

func extractBangumi(url, html string, extractOption types.Options) ([]*types.Data, error) {
	dataString := utils.MatchOneOf(html, `window.__INITIAL_STATE__=(.+?);\(function`)[1]
	var data bangumiData
	err := json.Unmarshal([]byte(dataString), &data)
	if err != nil {
		return nil, err
	}
	if !extractOption.Playlist {
		options := bilibiliOptions{
			url:     url,
			html:    html,
			bangumi: true,
			aid:     data.EpInfo.Aid,
			cid:     data.EpInfo.Cid,
		}
		return []*types.Data{bilibiliDownload(options, extractOption)}, nil
	}

	// handle bangumi playlist
	needDownloadItems := utils.NeedDownloadList(extractOption.Items, extractOption.ItemStart, extractOption.ItemEnd, len(data.EpList))
	extractedData := make([]*types.Data, len(needDownloadItems))
	wgp := utils.NewWaitGroupPool(extractOption.ThreadNumber)
	dataIndex := 0
	for index, u := range data.EpList {
		if !utils.ItemInSlice(index+1, needDownloadItems) {
			continue
		}
		wgp.Add()
		id := u.EpID
		if id == 0 {
			id = u.ID
		}
		// html content can't be reused here
		options := bilibiliOptions{
			url:     fmt.Sprintf("https://www.bilibili.com/bangumi/play/ep%d", id),
			bangumi: true,
			aid:     u.Aid,
			cid:     u.Cid,
		}
		go func(index int, options bilibiliOptions, extractedData []*types.Data) {
			defer wgp.Done()
			extractedData[index] = bilibiliDownload(options, extractOption)
		}(dataIndex, options, extractedData)
		dataIndex++
	}
	wgp.Wait()
	return extractedData, nil
}

func getMultiPageData(html string) (*multiPage, error) {
	var data multiPage
	multiPageDataString := utils.MatchOneOf(
		html, `window.__INITIAL_STATE__=(.+?);\(function`,
	)
	if multiPageDataString == nil {
		return &data, errors.New("this page has no playlist")
	}
	err := json.Unmarshal([]byte(multiPageDataString[1]), &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func extractNormalVideo(url, html string, extractOption types.Options) ([]*types.Data, error) {
	pageData, err := getMultiPageData(html)
	if err != nil {
		return nil, err
	}
	if !extractOption.Playlist {
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

		if len(pageData.VideoData.Pages) < p || p < 1 {
			return nil, types.ErrURLParseFailed
		}

		page := pageData.VideoData.Pages[p-1]
		options := bilibiliOptions{
			url:  url,
			html: html,
			aid:  pageData.Aid,
			cid:  page.Cid,
			page: p,
		}
		// "part":"" or "part":"Untitled"
		if page.Part == "Untitled" || len(pageData.VideoData.Pages) == 1 {
			options.subtitle = ""
		} else {
			options.subtitle = page.Part
		}
		return []*types.Data{bilibiliDownload(options, extractOption)}, nil
	}

	// handle normal video playlist
	// https://www.bilibili.com/video/av20827366/?p=1
	needDownloadItems := utils.NeedDownloadList(extractOption.Items, extractOption.ItemStart, extractOption.ItemEnd, len(pageData.VideoData.Pages))
	extractedData := make([]*types.Data, len(needDownloadItems))
	wgp := utils.NewWaitGroupPool(extractOption.ThreadNumber)
	dataIndex := 0
	for index, u := range pageData.VideoData.Pages {
		if !utils.ItemInSlice(index+1, needDownloadItems) {
			continue
		}
		wgp.Add()
		options := bilibiliOptions{
			url:      url,
			html:     html,
			aid:      pageData.Aid,
			cid:      u.Cid,
			subtitle: u.Part,
			page:     u.Page,
		}
		go func(index int, options bilibiliOptions, extractedData []*types.Data) {
			defer wgp.Done()
			extractedData[index] = bilibiliDownload(options, extractOption)
		}(dataIndex, options, extractedData)
		dataIndex++
	}
	wgp.Wait()
	return extractedData, nil
}

type extractor struct{}

// New returns a youtube extractor.
func New() types.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option types.Options) ([]*types.Data, error) {
	var err error
	html, err := request.Get(url, referer, nil)
	if err != nil {
		return nil, err
	}
	if strings.Contains(url, "bangumi") {
		// handle bangumi
		return extractBangumi(url, html, option)
	}
	// handle normal video
	return extractNormalVideo(url, html, option)
}

// bilibiliDownload is the download function for a single URL
func bilibiliDownload(options bilibiliOptions, extractOption types.Options) *types.Data {
	var (
		err        error
		html       string
		seasonType string
	)
	if options.html != "" {
		// reuse html string, but this can't be reused in case of playlist
		html = options.html
	} else {
		html, err = request.Get(options.url, referer, nil)
		if err != nil {
			return types.EmptyData(options.url, err)
		}
	}
	if options.bangumi {
		seasonType = utils.MatchOneOf(html, `"season_type":(\d+)`, `"ssType":(\d+)`)[1]
	}

	// Get "accept_quality" and "accept_description"
	// "accept_description":["高清 1080P","高清 720P","清晰 480P","流畅 360P"],
	// "accept_quality":[80,48,32,16],
	api, err := genAPI(options.aid, options.cid, options.bangumi, "15", seasonType, extractOption.Cookie)
	if err != nil {
		return types.EmptyData(options.url, err)
	}
	jsonString, err := request.Get(api, referer, nil)
	if err != nil {
		return types.EmptyData(options.url, err)
	}
	var quality qualityInfo
	err = json.Unmarshal([]byte(jsonString), &quality)
	if err != nil {
		return types.EmptyData(options.url, err)
	}

	streams := make(map[string]*types.Stream, len(quality.Quality))
	for _, q := range quality.Quality {
		apiURL, err := genAPI(options.aid, options.cid, options.bangumi, strconv.Itoa(q), seasonType, extractOption.Cookie)
		if err != nil {
			return types.EmptyData(options.url, err)
		}
		jsonString, err := request.Get(apiURL, referer, nil)
		if err != nil {
			return types.EmptyData(options.url, err)
		}
		var data bilibiliData
		err = json.Unmarshal([]byte(jsonString), &data)
		if err != nil {
			return types.EmptyData(options.url, err)
		}

		// Avoid duplicate streams
		if _, ok := streams[strconv.Itoa(data.Quality)]; ok {
			continue
		}

		parts, size := genParts(data.DURL)
		streams[strconv.Itoa(data.Quality)] = &types.Stream{
			Parts:   parts,
			Size:    size,
			Quality: qualityString[data.Quality],
		}
	}

	// get the title
	doc, err := parser.GetDoc(html)
	if err != nil {
		return types.EmptyData(options.url, err)
	}
	title := parser.Title(doc)
	if options.subtitle != "" {
		if extractOption.EpisodeTitleOnly {
			title = fmt.Sprintf("P%d %s", options.page, options.subtitle)
		} else {
			title = fmt.Sprintf("%s P%d %s", title, options.page, options.subtitle)
		}
	}

	return &types.Data{
		Site:    "哔哩哔哩 bilibili.com",
		Title:   title,
		Type:    types.DataTypeVideo,
		Streams: streams,
		Caption: &types.Part{
			URL: fmt.Sprintf("https://comment.bilibili.com/%d.xml", options.cid),
			Ext: "xml",
		},
		URL: options.url,
	}
}
