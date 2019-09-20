package douyu

import (
	"encoding/json"
	"errors"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/extractors"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type douyuData struct {
	Error int `json:"error"`
	Data  struct {
		VideoURL string `json:"video_url"`
	} `json:"data"`
}

type douyuURLInfo struct {
	URL  string
	Size int64
}

func douyuM3u8(url string) ([]douyuURLInfo, int64, error) {
	var (
		data            []douyuURLInfo
		temp            douyuURLInfo
		size, totalSize int64
		err             error
	)
	urls, err := utils.M3u8URLs(url)
	if err != nil {
		return nil, 0, err
	}
	for _, u := range urls {
		size, err = request.Size(u, url)
		if err != nil {
			return nil, 0, err
		}
		totalSize += size
		temp = douyuURLInfo{
			URL:  u,
			Size: size,
		}
		data = append(data, temp)
	}
	return data, totalSize, nil
}

// Extract is the main function for extracting data
func Extract(url string) ([]downloader.Data, error) {
	var err error
	liveVid := utils.MatchOneOf(url, `https?://www.douyu.com/(\S+)`)
	if liveVid != nil {
		return nil, errors.New("暂不支持斗鱼直播")
	}

	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, err
	}
	titles := utils.MatchOneOf(html, `<title>(.*?)</title>`)
	if titles == nil || len(titles) < 2 {
		return nil, extractors.ErrURLParseFailed
	}
	title := titles[1]

	vids := utils.MatchOneOf(url, `https?://v.douyu.com/show/(\S+)`)
	if vids == nil || len(vids) < 2 {
		return nil, extractors.ErrURLParseFailed
	}
	vid := vids[1]

	dataString, err := request.Get("http://vmobile.douyu.com/video/getInfo?vid="+vid, url, nil)
	if err != nil {
		return nil, err
	}
	var dataDict douyuData
	json.Unmarshal([]byte(dataString), &dataDict)

	m3u8URLs, totalSize, err := douyuM3u8(dataDict.Data.VideoURL)
	if err != nil {
		return nil, err
	}
	urls := make([]downloader.URL, len(m3u8URLs))
	for index, u := range m3u8URLs {
		urls[index] = downloader.URL{
			URL:  u.URL,
			Size: u.Size,
			Ext:  "ts",
		}
	}

	streams := map[string]downloader.Stream{
		"default": {
			URLs: urls,
			Size: totalSize,
		},
	}
	return []downloader.Data{
		{
			Site:    "斗鱼 douyu.com",
			Title:   title,
			Type:    "video",
			Streams: streams,
			URL:     url,
		},
	}, nil
}
