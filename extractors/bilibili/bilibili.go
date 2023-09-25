package bilibili

// import (
// 	"encoding/json"
// 	"fmt"
// 	"sort"
// 	"strconv"
// 	"strings"

// 	"github.com/pkg/errors"
// 	"github.com/wujiu2020/lux/extractors/proto"
// 	"github.com/wujiu2020/lux/parser"
// 	"github.com/wujiu2020/lux/request"
// 	"github.com/wujiu2020/lux/utils"
// )

// const (
// 	bilibiliAPI        = "https://api.bilibili.com/x/player/playurl?"
// 	bilibiliBangumiAPI = "https://api.bilibili.com/pgc/player/web/playurl?"
// 	bilibiliTokenAPI   = "https://api.bilibili.com/x/player/playurl/token?"
// )

// const referer = "https://www.bilibili.com"

// type bilibiliOptions struct {
// 	url      string
// 	html     string
// 	bangumi  bool
// 	aid      int
// 	cid      int
// 	bvid     string
// 	page     int
// 	subtitle string
// }

// // New returns a bilibili extractor.
// func New() proto.Extractor {
// 	return &extractor{}
// }

// type extractor struct{}

// // Extract is the main function to extract the data.
// func (e *extractor) Extract(url string) (*proto.Data, error) {
// 	var err error
// 	html, err := request.Get(url, referer, nil)
// 	if err != nil {
// 		return nil, errors.WithStack(err)
// 	}

// 	if strings.Contains(url, "bangumi") {
// 		// handle bangumi
// 		return extractBangumi(url, html)
// 	}
// 	// handle normal video
// 	return extractNormalVideo(url, html)
// }

// func extractBangumi(url, html string) (*proto.Data, error) {
// 	dataString := utils.MatchOneOf(html, `<script\s+id="__NEXT_DATA__"\s+type="application/json"\s*>(.*?)</script\s*>`)[1]
// 	epMapString := utils.MatchOneOf(dataString, `"epMap"\s*:\s*(.+?)\s*,\s*"initEpList"`)[1]
// 	fullVideoIdString := utils.MatchOneOf(dataString, `"videoId"\s*:\s*"(ep|ss)(\d+)"`)
// 	epSsString := fullVideoIdString[1] // "ep" or "ss"
// 	videoIdString := fullVideoIdString[2]

// 	var epMap map[string]json.RawMessage
// 	err := json.Unmarshal([]byte(epMapString), &epMap)
// 	if err != nil {
// 		return nil, errors.WithStack(err)
// 	}
// 	var data bangumiData
// 	for idString, jsonByte := range epMap {
// 		var epInfo bangumiEpData
// 		err := json.Unmarshal(jsonByte, &epInfo)
// 		if err != nil {
// 			return nil, errors.WithStack(err)
// 		}
// 		epID, err := strconv.ParseInt(idString, 10, 0)
// 		if err != nil {
// 			return nil, errors.WithStack(err)
// 		}
// 		epInfo.EpID = int(epID)
// 		if idString == videoIdString || (epSsString == "ss" && epInfo.TitleFormat == "第1话") {
// 			data.EpInfo = epInfo
// 		}
// 		data.EpList = append(data.EpList, epInfo)
// 	}

// 	sort.Slice(data.EpList, func(i, j int) bool {
// 		return data.EpList[i].EpID < data.EpList[j].EpID
// 	})

// 	aid := data.EpInfo.Aid
// 	cid := data.EpInfo.Cid
// 	bvid := data.EpInfo.BVid
// 	titleFormat := data.EpInfo.TitleFormat
// 	longTitle := data.EpInfo.LongTitle
// 	if aid <= 0 || cid <= 0 || bvid == "" {
// 		aid = data.EpList[0].Aid
// 		cid = data.EpList[0].Cid
// 		bvid = data.EpList[0].BVid
// 		titleFormat = data.EpList[0].TitleFormat
// 		longTitle = data.EpList[0].LongTitle
// 	}
// 	options := bilibiliOptions{
// 		url:     url,
// 		html:    html,
// 		bangumi: true,
// 		aid:     aid,
// 		cid:     cid,
// 		bvid:    bvid,

// 		subtitle: fmt.Sprintf("%s %s", titleFormat, longTitle),
// 	}
// 	return bilibiliDownload(options), nil
// }

// func extractNormalVideo(url, html string) (*proto.Data, error) {
// 	pageData, err := getMultiPageData(html)
// 	if err != nil {
// 		return nil, errors.WithStack(err)
// 	}
// 	// handle URL that has a playlist, mainly for unified titles
// 	// <h1> tag does not include subtitles
// 	// bangumi doesn't need this
// 	pageString := utils.MatchOneOf(url, `\?p=(\d+)`)
// 	var p int
// 	if pageString == nil {
// 		// https://www.bilibili.com/video/av20827366/
// 		p = 1
// 	} else {
// 		// https://www.bilibili.com/video/av20827366/?p=2
// 		p, _ = strconv.Atoi(pageString[1])
// 	}

// 	if len(pageData.VideoData.Pages) < p || p < 1 {
// 		return nil, errors.WithStack(proto.ErrURLParseFailed)
// 	}

// 	page := pageData.VideoData.Pages[p-1]
// 	options := bilibiliOptions{
// 		url:  url,
// 		html: html,
// 		aid:  pageData.Aid,
// 		bvid: pageData.BVid,
// 		cid:  page.Cid,
// 		page: p,
// 	}
// 	// "part":"" or "part":"Untitled"
// 	if page.Part == "Untitled" || len(pageData.VideoData.Pages) == 1 {
// 		options.subtitle = ""
// 	} else {
// 		options.subtitle = page.Part
// 	}
// 	return bilibiliDownload(options), nil
// }

// func bilibiliDownload(options bilibiliOptions) *proto.Data {
// 	var (
// 		err  error
// 		html string
// 	)
// 	if options.html != "" {
// 		// reuse html string, but this can't be reused in case of playlist
// 		html = options.html
// 	} else {
// 		html, err = request.Get(options.url, referer, nil)
// 		if err != nil {
// 			return proto.EmptyData(options.url, err)
// 		}
// 	}

// 	// Get "accept_quality" and "accept_description"
// 	// "accept_description":["超高清 8K","超清 4K","高清 1080P+","高清 1080P","高清 720P","清晰 480P","流畅 360P"],
// 	// "accept_quality":[127，120,112,80,48,32,16],
// 	api, err := genAPI(options.aid, options.cid, 127, options.bvid, options.bangumi)
// 	if err != nil {
// 		return proto.EmptyData(options.url, err)
// 	}
// 	jsonString, err := request.Get(api, referer, nil)
// 	if err != nil {
// 		return proto.EmptyData(options.url, err)
// 	}

// 	var data dash
// 	err = json.Unmarshal([]byte(jsonString), &data)
// 	if err != nil {
// 		return proto.EmptyData(options.url, err)
// 	}
// 	var dashData dashInfo
// 	if data.Data.Description == nil {
// 		dashData = data.Result
// 	} else {
// 		dashData = data.Data
// 	}

// 	var audioPart *proto.Part
// 	if dashData.Streams.Audio != nil {
// 		// Get audio part
// 		var audioID int
// 		audios := map[int]string{}
// 		bandwidth := 0
// 		for _, stream := range dashData.Streams.Audio {
// 			if stream.Bandwidth > bandwidth {
// 				audioID = stream.ID
// 				bandwidth = stream.Bandwidth
// 			}
// 			audios[stream.ID] = stream.BaseURL
// 		}
// 		s, err := request.Size(audios[audioID], referer)
// 		if err != nil {
// 			return proto.EmptyData(options.url, err)
// 		}
// 		audioPart = &proto.Part{
// 			URL:  audios[audioID],
// 			Size: s,
// 			Ext:  "m4a",
// 		}
// 	}

// 	streams := make([]*proto.Stream, 0)
// 	for _, stream := range dashData.Streams.Video {
// 		s, err := request.Size(stream.BaseURL, referer)
// 		if err != nil {
// 			return proto.EmptyData(options.url, err)
// 		}
// 		segs := make([]*proto.Part, 0, 2)
// 		segs = append(segs, &proto.Part{
// 			URL:  stream.BaseURL,
// 			Size: s,
// 			Ext:  getExtFromMimeType(stream.MimeType),
// 		})
// 		if audioPart != nil {
// 			segs = append(segs, audioPart)
// 		}
// 		var size int64
// 		for _, part := range segs {
// 			size += part.Size
// 		}
// 		id := fmt.Sprintf("%d-%d", stream.ID, stream.Codecid)
// 		streams = append(streams, &proto.Stream{
// 			ID:      id,
// 			Segs:    segs,
// 			Size:    size,
// 			Quality: fmt.Sprintf("%s %s", qualityString[stream.ID], stream.Codecs),
// 		})
// 	}

// 	for _, durl := range dashData.DURLs {
// 		var ext string
// 		switch dashData.DURLFormat {
// 		case "flv", "flv480":
// 			ext = "flv"
// 		case "mp4", "hdmp4": // nolint
// 			ext = "mp4"
// 		}

// 		segs := make([]*proto.Part, 0, 1)
// 		segs = append(segs, &proto.Part{
// 			URL:  durl.URL,
// 			Size: durl.Size,
// 			Ext:  ext,
// 		})

// 		streams = append(streams, &proto.Stream{
// 			ID:      strconv.Itoa(dashData.CurQuality),
// 			Segs:    segs,
// 			Size:    durl.Size,
// 			Quality: qualityString[dashData.CurQuality],
// 		})
// 	}

// 	// get the title
// 	doc, err := parser.GetDoc(html)
// 	if err != nil {
// 		return proto.EmptyData(options.url, err)
// 	}
// 	title := parser.Title(doc)
// 	if options.subtitle != "" {
// 		pageString := ""
// 		if options.page > 0 {
// 			pageString = fmt.Sprintf("P%d ", options.page)
// 		}
// 		title = fmt.Sprintf("%s %s%s", title, pageString, options.subtitle)
// 	}

// 	return &proto.Data{
// 		Site:    "哔哩哔哩 bilibili.com",
// 		Title:   title,
// 		Type:    proto.DataTypeVideo,
// 		Streams: streams,
// 		URL:     options.url,
// 	}
// }

// func genAPI(aid, cid, quality int, bvid string, bangumi bool) (string, error) {
// 	var (
// 		baseAPIURL string
// 		params     string
// 	)
// 	var api string
// 	if bangumi {
// 		// The parameters need to be sorted by name
// 		// qn=0 flag makes the CDN address different every time
// 		// quality=120(4k) is the highest quality so far
// 		params = fmt.Sprintf(
// 			"cid=%d&bvid=%s&qn=%d&type=&otype=json&fourk=1&fnver=0&fnval=16",
// 			cid, bvid, quality,
// 		)
// 		baseAPIURL = bilibiliBangumiAPI
// 	} else {
// 		params = fmt.Sprintf(
// 			"avid=%d&cid=%d&bvid=%s&qn=%d&type=&otype=json&fourk=1&fnver=0&fnval=2000",
// 			aid, cid, bvid, quality,
// 		)
// 		baseAPIURL = bilibiliAPI
// 	}
// 	api = baseAPIURL + params
// 	// bangumi utoken also need to put in params to sign, but the ordinary video doesn't need
// 	return api, nil
// }

// func getMultiPageData(html string) (*multiPage, error) {
// 	var data multiPage
// 	multiPageDataString := utils.MatchOneOf(
// 		html, `window.__INITIAL_STATE__=(.+?);\(function`,
// 	)
// 	if multiPageDataString == nil {
// 		return &data, errors.New("this page has no playlist")
// 	}
// 	err := json.Unmarshal([]byte(multiPageDataString[1]), &data)
// 	if err != nil {
// 		return nil, errors.WithStack(err)
// 	}
// 	return &data, nil
// }

// func getExtFromMimeType(mimeType string) string {
// 	exts := strings.Split(mimeType, "/")
// 	if len(exts) == 2 {
// 		return exts[1]
// 	}
// 	return "mp4"
// }
