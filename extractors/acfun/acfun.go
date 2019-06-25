package acfun

import (
	"crypto/rc4"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

func getVid(html string, bangumi bool) (string, string, error) {
	var (
		err   error
		title string
		vid   string
	)
	if bangumi {
		pageInfo := utils.MatchOneOf(html, `<script>var\ pageInfo\ =\ (.+)`)
		if pageInfo == nil {
			return "", "", errors.New("未能找到bangumi视频ID")
		}
		var bangumiPageInfo bangumiPageInfoData
		err = json.Unmarshal([]byte(pageInfo[1]), &bangumiPageInfo)
		if err != nil {
			return "", "", err
		}
		title = bangumiPageInfo.Album.Title + bangumiPageInfo.Video.Videos[0].NewTitle + bangumiPageInfo.Video.Videos[0].EpisodeName
		vid = strconv.Itoa(bangumiPageInfo.Video.Videos[0].VideoId)
		return title, vid, err
	}

	videoInfo := utils.MatchOneOf(html, `<script>window.pageInfo\ =\ window.videoInfo\ =\ (.*);`)
	if videoInfo == nil {
		return "", "", errors.New("未能找到视频ID")
	}
	var normalPageInfo normalPageInfoData
	err = json.Unmarshal([]byte(videoInfo[1]), &normalPageInfo)
	title = normalPageInfo.Title + normalPageInfo.CurrentVideoInfo.Title
	vid = normalPageInfo.CurrentVideoInfo.Id
	return title, vid, err
}

func getAPI(vid string, url string) (string, string, error) {
	var (
		err      error
		apiInfo  apiInfoData
		embsig   string
		sourceId string
	)
	headers := map[string]string{
		"deviceType": "2",
	}
	api, err := request.Get(fmt.Sprintf("http://api.aixifan.com/plays/youku/%s", vid), url, headers)
	if err != nil {
		return "", "", err
	}
	err = json.Unmarshal([]byte(api), &apiInfo)
	if err != nil {
		return "", "", err
	}
	if apiInfo.Code != 200 {
		return "", "", errors.New("解析视频source id失败")
	}
	embsig = apiInfo.Data.Embsig
	sourceId = apiInfo.Data.SourceId
	return embsig, sourceId, nil
}

func getEncryptedText(embsig string, sourceId string, url string) (string, error) {
	var (
		err           error
		encryptedInfo encryptedInfoData
	)
	apiUrl := fmt.Sprintf("http://player.acfun.cn/js_data?vid=%s&ct=86&ev=4&sign=%s&time=%d",
		sourceId, embsig, time.Now().Unix()*1000)
	api, err := request.Get(apiUrl, url, nil)
	err = json.Unmarshal([]byte(api), &encryptedInfo)
	if err != nil {
		return "", err
	}
	if encryptedInfo.E.Code != 0 {
		return "", errors.New("获取加密数据失败")
	}
	return encryptedInfo.Data, nil

}

func decryptText(data string) (string, error) {
	c, err := rc4.NewCipher([]byte("m1uN9G6cKz0mooZM"))
	if err != nil {
		return "", nil
	}
	src, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", nil
	}
	dst := make([]byte, len(src))
	c.XORKeyStream(dst, src)
	return string(dst), nil
}

func getExt(url string) string {
	matches := utils.MatchOneOf(url, `(flv|mp4)&?`)
	if matches == nil {
		return "unknown_video"
	}
	return matches[1]
}

// Extract is the main function for extracting data
func Extract(url string) ([]downloader.Data, error) {
	config.FakeHeaders["User-Agent"] = "Transmission/2.77"
	var (
		err   error
		title string
		vid   string
	)
	matches := utils.MatchOneOf(url, `https?://[^\.]*\.*acfun\.[^\.]+/(\D|bangumi)/\D\D(\d+)`)
	if matches == nil {
		return downloader.EmptyList, errors.New("地址有误")
	}

	html, err := request.Get(url, url, nil)
	if err != nil {
		return downloader.EmptyList, err
	}

	bangumi := utils.MatchOneOf(url, `https?://[^\.]*\.*acfun\.[^\.]+/bangumi/ab(\d+)`)
	title, vid, err = getVid(html, bangumi != nil)
	if err != nil {
		return downloader.EmptyList, err
	}

	embsig, sourceId, err := getAPI(vid, url)
	if err != nil {
		return downloader.EmptyList, err
	}

	encryptedText, err := getEncryptedText(embsig, sourceId, url)
	data, err := decryptText(encryptedText)
	var acfunStreams acfunStreamsData
	err = json.Unmarshal([]byte(data), &acfunStreams)
	if err != nil {
		return downloader.EmptyList, err
	}
	streams := map[string]downloader.Stream{}
	for _, acfunStream := range acfunStreams.Stream {
		if strings.HasPrefix(acfunStream.StreamType, "m3u8") {
			continue
		}
		urls := make([]downloader.URL, acfunStream.SliceNum)
		for index, u := range acfunStream.Segs {
			ext := getExt(u.Url)
			urls[index] = downloader.URL{
				URL:  u.Url,
				Size: u.Size,
				Ext:  ext,
			}
		}
		streams[acfunStream.StreamType] = downloader.Stream{
			URLs:    urls,
			Quality: acfunStream.Resolution,
			Size:    acfunStream.TotalSize,
		}
	}

	return []downloader.Data{
		{
			Site:    "AcFun acfun.cn",
			Title:   title,
			Type:    "video",
			Streams: streams,
			URL:     url,
		},
	}, nil
}
