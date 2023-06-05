package weibo

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	netURL "net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("weibo", New())
}

type playInfo struct {
	Title string            `json:"title"`
	URLs  map[string]string `json:"urls"`
}

type playData struct {
	PlayInfo playInfo `json:"Component_Play_Playinfo"`
}

type weiboData struct {
	Code string   `json:"code"`
	Data playData `json:"data"`
	Msg  string   `json:"msg"`
}

func getXSRFToken() (string, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	url := "https://weibo.com/ajax/getversion"
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.116 Safari/537.36")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close() // nolint

	cookie := res.Header.Get("Set-Cookie")
	if cookie == "" {
		return "", nil
	}
	xsrfTokens := utils.MatchOneOf(cookie, `XSRF-TOKEN=(.+?);`)
	if xsrfTokens == nil || len(xsrfTokens) != 2 {
		return "", nil
	}
	return xsrfTokens[1], nil
}

func downloadWeiboVideo(url string) ([]*extractors.Data, error) {
	urldata, err := netURL.Parse(url)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	api := fmt.Sprintf(
		"https://video.h5.weibo.cn/s/video/object?object_id=%s&mid=%s",
		strings.Split(urldata.Path, "/")[1], strings.Split(urldata.Path, "/")[2],
	)
	jsonString, err := request.Get(api, "", nil)

	if err != nil {
		return nil, errors.WithStack(err)
	}
	rawSummary := utils.MatchOneOf(jsonString, `"summary":"(.+?)",`)[1]
	summary, err := strconv.Unquote(strings.Replace(strconv.Quote(rawSummary), `\\u`, `\u`, -1))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	rawhdURL := utils.MatchOneOf(jsonString, `"hd_url":"([^"]+)",`)[1]
	unescapedhdURL, err := strconv.Unquote(strings.Replace(strconv.Quote(rawhdURL), `\\u`, `\u`, -1))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	realhdURL := strings.ReplaceAll(unescapedhdURL, `\/`, `/`)
	hdsize, err := request.Size(realhdURL, "")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	streams := make(map[string]*extractors.Stream, 2)
	streams["hd"] = &extractors.Stream{
		Parts: []*extractors.Part{
			{
				URL:  realhdURL,
				Size: hdsize,
				Ext:  "mp4",
			},
		},
		Size:    hdsize,
		Quality: "hd",
	}
	rawURL := utils.MatchOneOf(jsonString, `"url":"([^"]+)",`)[1]
	unescapedURL, err := strconv.Unquote(strings.Replace(strconv.Quote(rawURL), `\\u`, `\u`, -1))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	realURL := strings.ReplaceAll(unescapedURL, `\/`, `/`)
	size, err := request.Size(realURL, "")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	streams["sd"] = &extractors.Stream{
		Parts: []*extractors.Part{
			{
				URL:  realhdURL,
				Size: size,
				Ext:  "mp4",
			},
		},
		Size:    size,
		Quality: "sd",
	}
	return []*extractors.Data{
		{
			Site:    "微博 weibo.com",
			Title:   summary,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}

func downloadWeiboTV(url string) ([]*extractors.Data, error) {
	APIEndpoint := "https://weibo.com/tv/api/component?page="
	urldata, err := netURL.Parse(url)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	APIURL := APIEndpoint + netURL.QueryEscape(urldata.Path)
	token, err := getXSRFToken()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	headers := map[string]string{
		"Cookie":       "SUB=_2AkMpogLYf8NxqwJRmP0XxG7kbo10ww_EieKf_vMDJRMxHRl-yj_nqm4NtRB6AiIsKFFGRY4-UuGD5B1-Kf9glz3sp7Ii",
		"Referer":      utils.MatchOneOf(url, `^([^?]+)`)[1],
		"content-type": `application/x-www-form-urlencoded`,
	}
	if token != "" {
		headers["Cookie"] += "; XSRF-TOKEN=" + token
		headers["x-xsrf-token"] = token
	}
	oid := utils.MatchOneOf(url, `tv/show/([^?]+)`)[1]
	postData := "data=" + netURL.QueryEscape("{\"Component_Play_Playinfo\":{\"oid\":\""+oid+"\"}}")
	payload := strings.NewReader(postData)
	res, err := request.Request(http.MethodPost, APIURL, payload, headers)

	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer res.Body.Close() // nolint
	var dataReader io.ReadCloser
	if res.Header.Get("Content-Encoding") == "gzip" {
		dataReader, err = gzip.NewReader(res.Body)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	} else {
		dataReader = res.Body
	}
	var data weiboData
	if err = json.NewDecoder(dataReader).Decode(&data); err != nil {
		return nil, errors.WithStack(err)
	}

	if data.Data.PlayInfo.URLs == nil {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}
	realURLs := map[string]string{}
	for k, v := range data.Data.PlayInfo.URLs {
		if strings.HasPrefix(v, "http") {
			continue
		}
		realURLs[k] = "https:" + v
	}

	streams := make(map[string]*extractors.Stream, len(realURLs))
	for q, u := range realURLs {
		size, err := request.Size(u, "")
		if err != nil {
			return nil, errors.WithStack(err)
		}
		streams[q] = &extractors.Stream{
			Parts: []*extractors.Part{
				{
					URL:  u,
					Size: size,
					Ext:  "mp4",
				},
			},
			Size:    size,
			Quality: q,
		}
	}
	return []*extractors.Data{
		{
			Site:    "微博 weibo.com",
			Title:   data.Data.PlayInfo.Title,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}

type extractor struct{}

// New returns a weibo extractor.
func New() extractors.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	if !strings.Contains(url, "m.weibo.cn") {
		if strings.Contains(url, "weibo.com/tv/show/") {
			return downloadWeiboTV(url)
		} else if strings.Contains(url, "video.h5.weibo.cn") {
			return downloadWeiboVideo(url)
		}
		url = strings.Replace(url, "weibo.com", "m.weibo.cn", 1)
	}
	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	titles := utils.MatchOneOf(
		html, `"content2": "(.+?)",`, `"status_title": "(.+?)",`,
	)
	if titles == nil || len(titles) < 2 {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}
	title := titles[1]

	urlsJsonStrs := utils.MatchOneOf(
		html, `"urls": (\{[^\}]+\})`,
	)
	if urlsJsonStrs == nil || len(urlsJsonStrs) < 2 {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}
	urlsJson := urlsJsonStrs[1]
	var qualityUrls map[string]string
	err = json.Unmarshal([]byte(urlsJson), &qualityUrls)
	if err != nil {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}

	streams := make(map[string]*extractors.Stream)
	var size int64
	for quality, realURL := range qualityUrls {
		streamId := quality
		size, err = request.Size(realURL, url)
		if err != nil {
			continue
		}
		urlData := &extractors.Part{
			URL:  realURL,
			Size: size,
			Ext:  "mp4",
		}
		streams[streamId] = &extractors.Stream{
			Parts: []*extractors.Part{urlData},
			Size:  size,
		}
	}
	if err != nil || len(streams) <= 0 {
		return nil, errors.WithStack(err)
	}

	return []*extractors.Data{
		{
			Site:    "微博 weibo.com",
			Title:   title,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
