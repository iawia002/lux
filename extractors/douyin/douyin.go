package douyin

import (
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

// Download main download function
func Download(url string) ([]downloader.Data, error) {
	var err error
	html, err := request.Get(url, url, nil)
	if err != nil {
		return downloader.EmptyList, err
	}
	title := utils.MatchOneOf(html, `<p class="desc">(.+?)</p>`)[1]
	realURL := utils.MatchOneOf(html, `playAddr: "(.+?)"`)[1]
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
			Site:    "抖音 douyin.com",
			Title:   utils.FileName(title),
			Type:    "video",
			Streams: streams,
			URL:     url,
		},
	}, nil
}
