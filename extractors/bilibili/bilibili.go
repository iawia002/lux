package bilibili

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/iawia002/lia/array"
	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/parser"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("bilibili", New())
}

const (
	bilibiliAPI        = "https://api.bilibili.com/x/player/playurl?"
	bilibiliBangumiAPI = "https://api.bilibili.com/pgc/player/web/playurl?"
	bilibiliTokenAPI   = "https://api.bilibili.com/x/player/playurl/token?"
)

const referer = "https://www.bilibili.com"

var utoken string

func genAPI(aid, cid, quality int, bvid string, bangumi bool, cookie string) (string, error) {
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
			return "", errors.Errorf("cookie error: %s", t.Message)
		}
		utoken = t.Data.Token
	}
	var api string
	if bangumi {
		// The parameters need to be sorted by name
		// qn=0 flag makes the CDN address different every time
		// quality=120(4k) is the highest quality so far
		params = fmt.Sprintf(
			"cid=%d&bvid=%s&qn=%d&type=&otype=json&fourk=1&fnver=0&fnval=16",
			cid, bvid, quality,
		)
		baseAPIURL = bilibiliBangumiAPI
	} else {
		params = fmt.Sprintf(
			"avid=%d&cid=%d&bvid=%s&qn=%d&type=&otype=json&fourk=1&fnver=0&fnval=2000",
			aid, cid, bvid, quality,
		)
		baseAPIURL = bilibiliAPI
	}
	api = baseAPIURL + params
	// bangumi utoken also need to put in params to sign, but the ordinary video doesn't need
	if !bangumi && utoken != "" {
		api = fmt.Sprintf("%s&utoken=%s", api, utoken)
	}
	return api, nil
}

type bilibiliOptions struct {
	url      string
	html     string
	bangumi  bool
	aid      int
	cid      int
	bvid     string
	page     int
	subtitle string
}

func extractBangumi(url, html string, extractOption extractors.Options) ([]*extractors.Data, error) {
	dataString := utils.MatchOneOf(html, `<script\s+id="__NEXT_DATA__"\s+type="application/json"\s*>(.*?)</script\s*>`)[1]
	epMapString := utils.MatchOneOf(dataString, `"epMap"\s*:\s*(.+?)\s*,\s*"initEpList"`)[1]
	fullVideoIdString := utils.MatchOneOf(dataString, `"videoId"\s*:\s*"(ep|ss)(\d+)"`)
	epSsString := fullVideoIdString[1] // "ep" or "ss"
	videoIdString := fullVideoIdString[2]

	var epMap map[string]json.RawMessage
	err := json.Unmarshal([]byte(epMapString), &epMap)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var data bangumiData
	for idString, jsonByte := range epMap {
		var epInfo bangumiEpData
		err := json.Unmarshal(jsonByte, &epInfo)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		epID, err := strconv.ParseInt(idString, 10, 0)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		epInfo.EpID = int(epID)
		if idString == videoIdString || (epSsString == "ss" && epInfo.TitleFormat == "第1话") {
			data.EpInfo = epInfo
		}
		data.EpList = append(data.EpList, epInfo)
	}

	sort.Slice(data.EpList, func(i, j int) bool {
		return data.EpList[i].EpID < data.EpList[j].EpID
	})

	if !extractOption.Playlist {
		aid := data.EpInfo.Aid
		cid := data.EpInfo.Cid
		bvid := data.EpInfo.BVid
		titleFormat := data.EpInfo.TitleFormat
		longTitle := data.EpInfo.LongTitle
		if aid <= 0 || cid <= 0 || bvid == "" {
			aid = data.EpList[0].Aid
			cid = data.EpList[0].Cid
			bvid = data.EpList[0].BVid
			titleFormat = data.EpList[0].TitleFormat
			longTitle = data.EpList[0].LongTitle
		}
		options := bilibiliOptions{
			url:     url,
			html:    html,
			bangumi: true,
			aid:     aid,
			cid:     cid,
			bvid:    bvid,

			subtitle: fmt.Sprintf("%s %s", titleFormat, longTitle),
		}
		return []*extractors.Data{bilibiliDownload(options, extractOption)}, nil
	}

	// handle bangumi playlist
	needDownloadItems := utils.NeedDownloadList(extractOption.Items, extractOption.ItemStart, extractOption.ItemEnd, len(data.EpList))
	extractedData := make([]*extractors.Data, len(needDownloadItems))
	wgp := utils.NewWaitGroupPool(extractOption.ThreadNumber)
	dataIndex := 0
	for index, u := range data.EpList {
		if !array.ItemInArray(index+1, needDownloadItems) {
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
			bvid:    u.BVid,

			subtitle: fmt.Sprintf("%s %s", u.TitleFormat, u.LongTitle),
		}
		go func(index int, options bilibiliOptions, extractedData []*extractors.Data) {
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
		return nil, errors.WithStack(err)
	}
	return &data, nil
}

func extractNormalVideo(url, html string, extractOption extractors.Options) ([]*extractors.Data, error) {
	pageData, err := getMultiPageData(html)
	if err != nil {
		return nil, errors.WithStack(err)
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
			return nil, errors.WithStack(extractors.ErrURLParseFailed)
		}

		page := pageData.VideoData.Pages[p-1]
		options := bilibiliOptions{
			url:  url,
			html: html,
			aid:  pageData.Aid,
			bvid: pageData.BVid,
			cid:  page.Cid,
			page: p,
		}
		// "part":"" or "part":"Untitled"
		if page.Part == "Untitled" || len(pageData.VideoData.Pages) == 1 {
			options.subtitle = ""
		} else {
			options.subtitle = page.Part
		}
		return []*extractors.Data{bilibiliDownload(options, extractOption)}, nil
	}

	// handle normal video playlist
	if len(pageData.Sections) == 0 {
		// https://www.bilibili.com/video/av20827366/?p=* each video in playlist has different p=?
		return multiPageDownload(url, html, extractOption, pageData)
	}
	// handle another kind of playlist
	// https://www.bilibili.com/video/av*** each video in playlist has different av/bv id
	return multiEpisodeDownload(url, html, extractOption, pageData)
}

// handle multi episode download
func multiEpisodeDownload(url, html string, extractOption extractors.Options, pageData *multiPage) ([]*extractors.Data, error) {
	needDownloadItems := utils.NeedDownloadList(extractOption.Items, extractOption.ItemStart, extractOption.ItemEnd, len(pageData.Sections[0].Episodes))
	extractedData := make([]*extractors.Data, len(needDownloadItems))
	wgp := utils.NewWaitGroupPool(extractOption.ThreadNumber)
	dataIndex := 0
	for index, u := range pageData.Sections[0].Episodes {
		if !array.ItemInArray(index+1, needDownloadItems) {
			continue
		}
		wgp.Add()
		options := bilibiliOptions{
			url:      url,
			html:     html,
			aid:      u.Aid,
			bvid:     u.BVid,
			cid:      u.Cid,
			subtitle: u.Title,
		}
		go func(index int, options bilibiliOptions, extractedData []*extractors.Data) {
			defer wgp.Done()
			extractedData[index] = bilibiliDownload(options, extractOption)
		}(dataIndex, options, extractedData)
		dataIndex++
	}
	wgp.Wait()
	return extractedData, nil
}

// handle multi page download
func multiPageDownload(url, html string, extractOption extractors.Options, pageData *multiPage) ([]*extractors.Data, error) {
	needDownloadItems := utils.NeedDownloadList(extractOption.Items, extractOption.ItemStart, extractOption.ItemEnd, len(pageData.VideoData.Pages))
	extractedData := make([]*extractors.Data, len(needDownloadItems))
	wgp := utils.NewWaitGroupPool(extractOption.ThreadNumber)
	dataIndex := 0
	for index, u := range pageData.VideoData.Pages {
		if !array.ItemInArray(index+1, needDownloadItems) {
			continue
		}
		wgp.Add()
		options := bilibiliOptions{
			url:      url,
			html:     html,
			aid:      pageData.Aid,
			bvid:     pageData.BVid,
			cid:      u.Cid,
			subtitle: u.Part,
			page:     u.Page,
		}
		go func(index int, options bilibiliOptions, extractedData []*extractors.Data) {
			defer wgp.Done()
			extractedData[index] = bilibiliDownload(options, extractOption)
		}(dataIndex, options, extractedData)
		dataIndex++
	}
	wgp.Wait()
	return extractedData, nil
}

type extractor struct{}

// New returns a bilibili extractor.
func New() extractors.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	var err error
	html, err := request.Get(url, referer, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// set thread number to 1 manually to avoid http 412 error
	option.ThreadNumber = 1

	if strings.Contains(url, "bangumi") {
		// handle bangumi
		return extractBangumi(url, html, option)
	}
	// handle normal video
	return extractNormalVideo(url, html, option)
}

// bilibiliDownload is the download function for a single URL
func bilibiliDownload(options bilibiliOptions, extractOption extractors.Options) *extractors.Data {
	var (
		err  error
		html string
	)
	if options.html != "" {
		// reuse html string, but this can't be reused in case of playlist
		html = options.html
	} else {
		html, err = request.Get(options.url, referer, nil)
		if err != nil {
			return extractors.EmptyData(options.url, err)
		}
	}

	// Get "accept_quality" and "accept_description"
	// "accept_description":["超高清 8K","超清 4K","高清 1080P+","高清 1080P","高清 720P","清晰 480P","流畅 360P"],
	// "accept_quality":[127，120,112,80,48,32,16],
	api, err := genAPI(options.aid, options.cid, 127, options.bvid, options.bangumi, extractOption.Cookie)
	if err != nil {
		return extractors.EmptyData(options.url, err)
	}
	jsonString, err := request.Get(api, referer, nil)
	if err != nil {
		return extractors.EmptyData(options.url, err)
	}

	var data dash
	err = json.Unmarshal([]byte(jsonString), &data)
	if err != nil {
		return extractors.EmptyData(options.url, err)
	}
	var dashData dashInfo
	if data.Data.Description == nil {
		dashData = data.Result
	} else {
		dashData = data.Data
	}

	var audioPart *extractors.Part
	if dashData.Streams.Audio != nil {
		// Get audio part
		var audioID int
		audios := map[int]string{}
		bandwidth := 0
		for _, stream := range dashData.Streams.Audio {
			if stream.Bandwidth > bandwidth {
				audioID = stream.ID
				bandwidth = stream.Bandwidth
			}
			audios[stream.ID] = stream.BaseURL
		}
		s, err := request.Size(audios[audioID], referer)
		if err != nil {
			return extractors.EmptyData(options.url, err)
		}
		audioPart = &extractors.Part{
			URL:  audios[audioID],
			Size: s,
			Ext:  "m4a",
		}
	}

	streams := make(map[string]*extractors.Stream, len(dashData.Quality))
	for _, stream := range dashData.Streams.Video {
		s, err := request.Size(stream.BaseURL, referer)
		if err != nil {
			return extractors.EmptyData(options.url, err)
		}
		parts := make([]*extractors.Part, 0, 2)
		parts = append(parts, &extractors.Part{
			URL:  stream.BaseURL,
			Size: s,
			Ext:  getExtFromMimeType(stream.MimeType),
		})
		if audioPart != nil {
			parts = append(parts, audioPart)
		}
		var size int64
		for _, part := range parts {
			size += part.Size
		}
		id := fmt.Sprintf("%d-%d", stream.ID, stream.Codecid)
		streams[id] = &extractors.Stream{
			Parts:   parts,
			Size:    size,
			Quality: fmt.Sprintf("%s %s", qualityString[stream.ID], stream.Codecs),
		}
		if audioPart != nil {
			streams[id].NeedMux = true
		}
	}

	for _, durl := range dashData.DURLs {
		var ext string
		switch dashData.DURLFormat {
		case "flv", "flv480":
			ext = "flv"
		case "mp4", "hdmp4": // nolint
			ext = "mp4"
		}

		parts := make([]*extractors.Part, 0, 1)
		parts = append(parts, &extractors.Part{
			URL:  durl.URL,
			Size: durl.Size,
			Ext:  ext,
		})

		streams[strconv.Itoa(dashData.CurQuality)] = &extractors.Stream{
			Parts:   parts,
			Size:    durl.Size,
			Quality: qualityString[dashData.CurQuality],
		}
	}

	// get the title
	doc, err := parser.GetDoc(html)
	if err != nil {
		return extractors.EmptyData(options.url, err)
	}
	title := parser.Title(doc)
	if options.subtitle != "" {
		pageString := ""
		if options.page > 0 {
			pageString = fmt.Sprintf("P%d ", options.page)
		}
		if extractOption.EpisodeTitleOnly {
			title = fmt.Sprintf("%s%s", pageString, options.subtitle)
		} else {
			title = fmt.Sprintf("%s %s%s", title, pageString, options.subtitle)
		}
	}

	return &extractors.Data{
		Site:    "哔哩哔哩 bilibili.com",
		Title:   title,
		Type:    extractors.DataTypeVideo,
		Streams: streams,
		Captions: map[string]*extractors.CaptionPart{
			"danmaku": {
				Part: extractors.Part{
					URL: fmt.Sprintf("https://comment.bilibili.com/%d.xml", options.cid),
					Ext: "xml",
				},
			},
			"subtitle": getSubTitleCaptionPart(options.aid, options.cid),
		},
		URL: options.url,
	}
}

func getExtFromMimeType(mimeType string) string {
	exts := strings.Split(mimeType, "/")
	if len(exts) == 2 {
		return exts[1]
	}
	return "mp4"
}

func getSubTitleCaptionPart(aid int, cid int) *extractors.CaptionPart {
	jsonString, err := request.Get(
		fmt.Sprintf("http://api.bilibili.com/x/web-interface/view?aid=%d&cid=%d", aid, cid), referer, nil,
	)
	if err != nil {
		return nil
	}
	stu := bilibiliWebInterface{}
	err = json.Unmarshal([]byte(jsonString), &stu)
	if err != nil || len(stu.Data.SubtitleInfo.SubtitleList) == 0 {
		return nil
	}
	return &extractors.CaptionPart{
		Part: extractors.Part{
			URL: stu.Data.SubtitleInfo.SubtitleList[0].SubtitleUrl,
			Ext: "srt",
		},
		Transform: subtitleTransform,
	}
}

func subtitleTransform(body []byte) ([]byte, error) {
	bytes := ""
	captionData := bilibiliSubtitleFormat{}
	err := json.Unmarshal(body, &captionData)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	for i := 0; i < len(captionData.Body); i++ {
		bytes += fmt.Sprintf("%d\n%s --> %s\n%s\n\n",
			i,
			time.Unix(0, int64(captionData.Body[i].From*1000)*int64(time.Millisecond)).UTC().Format("15:04:05.000"),
			time.Unix(0, int64(captionData.Body[i].To*1000)*int64(time.Millisecond)).UTC().Format("15:04:05.000"),
			captionData.Body[i].Content,
		)
	}
	return []byte(bytes), nil
}
