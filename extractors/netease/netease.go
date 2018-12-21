package netease

import (
	"errors"
	netURL "net/url"
	"strings"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

// Extract is the main function for extracting data
func Extract(url string) ([]downloader.Data, error) {
	url = strings.Replace(url, "/#/", "/", 1)
	vid := utils.MatchOneOf(url, `/(mv|video)\?id=(\w+)`)
	if vid == nil {
		return downloader.EmptyList, errors.New("invalid url for netease music")
	}
	var err error
	html, err := request.Get(url, url, nil)
	if err != nil {
		return downloader.EmptyList, err
	}
	if strings.Contains(html, "u-errlg-404") {
		return downloader.EmptyList, errors.New("404 music not found")
	}
	title := utils.MatchOneOf(html, `<meta property="og:title" content="(.+?)" />`)[1]
	realURL := utils.MatchOneOf(html, `<meta property="og:video" content="(.+?)" />`)[1]
	realURL, _ = netURL.QueryUnescape(realURL)
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
			Site:    "网易云音乐 music.163.com",
			Title:   title,
			Type:    "video",
			Streams: streams,
			URL:     url,
		},
	}, nil
}
