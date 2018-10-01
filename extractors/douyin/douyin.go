package douyin

import (
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

// Download main download function
func Download(url string) ([]downloader.VideoData, error) {
	var err error
	html, err := request.Get(url, url, nil)
	if err != nil {
		return downloader.EmptyData, err
	}
	title := utils.MatchOneOf(html, `<p class="desc">(.+?)</p>`)[1]
	realURL := utils.MatchOneOf(html, `playAddr: "(.+?)"`)[1]
	size, err := request.Size(realURL, url)
	if err != nil {
		return downloader.EmptyData, err
	}
	urlData := downloader.URLData{
		URL:  realURL,
		Size: size,
		Ext:  "mp4",
	}
	format := map[string]downloader.FormatData{
		"default": {
			URLs: []downloader.URLData{urlData},
			Size: size,
		},
	}
	return []downloader.VideoData{
		{
			Site:    "抖音 douyin.com",
			Title:   utils.FileName(title),
			Type:    "video",
			Formats: format,
		},
	}, nil
}
