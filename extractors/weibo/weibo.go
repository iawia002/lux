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

	"github.com/iawia002/annie/extractors/types"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

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
	token := utils.MatchOneOf(res.Header.Get("Set-Cookie"), `XSRF-TOKEN=(.+?);`)[1]
	return token, nil
}

func downloadWeiboVideo(url string) ([]*types.Data, error) {
	urldata, err := netURL.Parse(url)
	if err != nil {
		return nil, err
	}
	api := fmt.Sprintf(
		"https://video.h5.weibo.cn/s/video/object?object_id=%s&mid=%s",
		strings.Split(urldata.Path, "/")[1], strings.Split(urldata.Path, "/")[2],
	)
	jsonString, err := request.Get(api, "", nil)

	if err != nil {
		return nil, err
	}
	rawSummary := utils.MatchOneOf(jsonString, `"summary":"(.+?)",`)[1]
	summary, err := strconv.Unquote(strings.Replace(strconv.Quote(rawSummary), `\\u`, `\u`, -1))
	if err != nil {
		return nil, err
	}
	rawhdURL := utils.MatchOneOf(jsonString, `"hd_url":"([^"]+)",`)[1]
	unescapedhdURL, err := strconv.Unquote(strings.Replace(strconv.Quote(rawhdURL), `\\u`, `\u`, -1))
	if err != nil {
		return nil, err
	}
	realhdURL := strings.ReplaceAll(unescapedhdURL, `\/`, `/`)
	hdsize, err := request.Size(realhdURL, "")
	if err != nil {
		return nil, err
	}
	streams := make(map[string]*types.Stream, 2)
	streams["hd"] = &types.Stream{
		Parts: []*types.Part{
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
		return nil, err
	}
	realURL := strings.ReplaceAll(unescapedURL, `\/`, `/`)
	size, err := request.Size(realURL, "")
	if err != nil {
		return nil, err
	}
	streams["sd"] = &types.Stream{
		Parts: []*types.Part{
			{
				URL:  realhdURL,
				Size: size,
				Ext:  "mp4",
			},
		},
		Size:    size,
		Quality: "sd",
	}
	return []*types.Data{
		{
			Site:    "微博 weibo.com",
			Title:   summary,
			Type:    types.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}

func downloadWeiboTV(url string) ([]*types.Data, error) {
	APIEndpoint := "https://weibo.com/tv/api/component?page="
	urldata, err := netURL.Parse(url)
	if err != nil {
		return nil, err
	}
	APIURL := APIEndpoint + netURL.QueryEscape(urldata.Path)
	token, err := getXSRFToken()
	if err != nil {
		return nil, err
	}
	headers := map[string]string{
		"Cookie":       "SUB=_2AkMpogLYf8NxqwJRmP0XxG7kbo10ww_EieKf_vMDJRMxHRl-yj_nqm4NtRB6AiIsKFFGRY4-UuGD5B1-Kf9glz3sp7Ii; XSRF-TOKEN=" + token,
		"Referer":      utils.MatchOneOf(url, `^([^?]+)`)[1],
		"content-type": `application/x-www-form-urlencoded`,
		"x-xsrf-token": token,
	}
	oid := utils.MatchOneOf(url, `tv/show/([^?]+)`)[1]
	postData := "data=" + netURL.QueryEscape("{\"Component_Play_Playinfo\":{\"oid\":\""+oid+"\"}}")
	payload := strings.NewReader(postData)
	res, err := request.Request(http.MethodPost, APIURL, payload, headers)

	if err != nil {
		return nil, err
	}
	defer res.Body.Close() // nolint
	var dataReader io.ReadCloser
	if res.Header.Get("Content-Encoding") == "gzip" {
		dataReader, err = gzip.NewReader(res.Body)
		if err != nil {
			return nil, err
		}
	} else {
		dataReader = res.Body
	}
	var data weiboData
	if err = json.NewDecoder(dataReader).Decode(&data); err != nil {
		return nil, err
	}

	if data.Data.PlayInfo.URLs == nil {
		return nil, types.ErrURLParseFailed
	}
	realURLs := map[string]string{}
	for k, v := range data.Data.PlayInfo.URLs {
		if strings.HasPrefix(v, "http") {
			continue
		}
		realURLs[k] = "https:" + v
	}

	streams := make(map[string]*types.Stream, len(realURLs))
	for q, u := range realURLs {
		size, err := request.Size(u, "")
		if err != nil {
			return nil, err
		}
		streams[q] = &types.Stream{
			Parts: []*types.Part{
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
	return []*types.Data{
		{
			Site:    "微博 weibo.com",
			Title:   data.Data.PlayInfo.Title,
			Type:    types.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}

type extractor struct{}

// New returns a youtube extractor.
func New() types.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option types.Options) ([]*types.Data, error) {
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
		return nil, err
	}
	titles := utils.MatchOneOf(
		html, `"content2": "(.+?)",`, `"status_title": "(.+?)",`,
	)
	if titles == nil || len(titles) < 2 {
		return nil, types.ErrURLParseFailed
	}
	title := titles[1]

	realURLs := utils.MatchOneOf(
		html, `"stream_url_hd": "(.+?)"`, `"stream_url": "(.+?)"`,
	)
	if realURLs == nil || len(realURLs) < 2 {
		return nil, types.ErrURLParseFailed
	}
	realURL := realURLs[1]

	size, err := request.Size(realURL, url)
	if err != nil {
		return nil, err
	}
	urlData := &types.Part{
		URL:  realURL,
		Size: size,
		Ext:  "mp4",
	}
	streams := map[string]*types.Stream{
		"default": {
			Parts: []*types.Part{urlData},
			Size:  size,
		},
	}

	return []*types.Data{
		{
			Site:    "微博 weibo.com",
			Title:   title,
			Type:    types.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
