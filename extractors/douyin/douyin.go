package douyin

import (
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/extractors"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

// Extract is the main function for extracting data
func Extract(url string) ([]downloader.Data, error) {
	var err error
	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, err
	}
	var title string
	desc := utils.MatchOneOf(html, `<p class="desc">(.+?)</p>`)
	if desc != nil {
		title = desc[1]
	} else {
		title = "抖音短视频"
	}
	realURLs := utils.MatchOneOf(html, `playAddr: "(.+?)"`)
	if realURLs == nil || len(realURLs) < 2 {
		return nil, extractors.ErrURLParseFailed
	}
	realURL := realURLs[1]

	size, err := request.Size(realURL, url)
	if err != nil {
		return nil, err
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
			Title:   title,
			Type:    "video",
			Streams: streams,
			URL:     url,
		},
	}, nil
}
