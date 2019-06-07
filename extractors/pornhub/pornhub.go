package pornhub

import (
	"encoding/json"
	"fmt"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type pornhubData struct {
	Format   string `json:"format"`
	Quality  string `json:"quality"`
	VideoURL string `json:"videoUrl"`
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

	streams := make(map[string]downloader.Stream, len(pornhubs))
	for _, data := range pornhubs {
		realURL := data.VideoURL
		if realURL == "" {
			continue
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
		streams[data.Quality] = downloader.Stream{
			URLs:    []downloader.URL{urlData},
			Size:    size,
			Quality: fmt.Sprintf("%sP", data.Quality),
		}
	}

	return []downloader.Data{
		{
			Site:    "Pornhub pornhub.com",
			Title:   title,
			Type:    "video",
			Streams: streams,
			URL:     url,
		},
	}, nil
}
