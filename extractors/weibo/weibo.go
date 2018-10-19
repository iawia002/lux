package weibo

import (
	"strings"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

// Download main download function
func Download(url string) ([]downloader.Data, error) {
	if !strings.Contains(url, "m.weibo.cn") {
		statusID := utils.MatchOneOf(url, `weibo\.com/tv/v/([^\?/]+)`)
		if statusID != nil {
			url = "https://m.weibo.cn/status/" + statusID[1]
		} else {
			url = strings.Replace(url, "weibo.com", "m.weibo.cn", 1)
		}
	}
	html, err := request.Get(url, url, nil)
	if err != nil {
		return downloader.EmptyList, err
	}
	title := utils.MatchOneOf(
		html, `"content2": "(.+?)",`, `"status_title": "(.+?)",`,
	)[1]
	realURL := utils.MatchOneOf(
		html, `"stream_url_hd": "(.+?)"`, `"stream_url": "(.+?)"`,
	)[1]
	size, err := request.Size(realURL, url)
	if err != nil {
		return downloader.EmptyList, err
	}
	urlData := downloader.URL{
		URL:  realURL,
		Size: size,
		Ext:  "mp4",
	}
	streams := map[string]downloader.Stream{
		"default": {
			URLs: []downloader.URL{urlData},
			Size: size,
		},
	}

	return []downloader.Data{
		{
			Site:    "微博 weibo.com",
			Title:   utils.FileName(title),
			Type:    "video",
			Streams: streams,
			URL:     url,
		},
	}, nil
}
