package pornhub

import (
	"encoding/json"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type pornhubData struct {
	DefaultQuality bool   `json:"defaultQuality"`
	Format         string `json:"format"`
	Quality        string `json:"quality"`
	VideoURL       string `json:"videoUrl"`
}

// Extract is the main function for extracting data
func Extract(url string) ([]downloader.Data, error) {
	var err error
	html, err := request.Get(url, url, nil)
	if err != nil {
		return downloader.EmptyList, err
	}
	var title string
	desc := utils.MatchOneOf(html, `<span class="inlineFree">(.+?)</span>`)
	if desc != nil {
		title = desc[1]
	} else {
		title = "pornhub video"
	}

	realURLs := utils.MatchOneOf(html, `"mediaDefinitions":(.+?),"isVertical"`)

	var pornhubs []pornhubData
	err = json.Unmarshal([]byte(realURLs[1]), &pornhubs)
	if err != nil {
		return downloader.EmptyList, err
	}

	//TODO add support for different quality
	var realURL string
	for _, downloadlink := range(pornhubs) {
		if downloadlink.VideoURL != "" {
			realURL = downloadlink.VideoURL
			break
		}
	}

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
			Site:    "Pornhub",
			Title:   title,
			Type:    "video",
			Streams: streams,
			URL:     url,
		},
	}, nil
}