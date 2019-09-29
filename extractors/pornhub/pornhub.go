package pornhub

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/extractors"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type pornhubData struct {
	Format   string          `json:"format"`
	Quality  json.RawMessage `json:"quality"`
	VideoURL string          `json:"videoUrl"`
}

// Extract is the main function for extracting data
func Extract(url string) ([]downloader.Data, error) {
	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, err
	}

	var title string
	desc := utils.MatchOneOf(html, `<span class="inlineFree">(.+?)</span>`)
	if desc != nil && len(desc) > 1 {
		title = desc[1]
	} else {
		title = "pornhub video"
	}

	realURLs := utils.MatchOneOf(html, `"mediaDefinitions":(.+?),"isVertical"`)
	if realURLs == nil || len(realURLs) < 2 {
		return nil, extractors.ErrURLParseFailed
	}

	var pornhubs []pornhubData
	if err = json.Unmarshal([]byte(realURLs[1]), &pornhubs); err != nil {
		return nil, err
	}

	streams := make(map[string]downloader.Stream, len(pornhubs))
	for _, data := range pornhubs {
		if data.Format == "hls" {
			continue
		}

		if bytes.ContainsRune(data.Quality, '[') {
			// skip the case where the quality value is an array
			// "quality": [
			//   720,
			//   480,
			//   240
			// ]
			continue
		}
		quality := string(data.Quality)

		realURL := data.VideoURL
		if len(realURL) == 0 {
			continue
		}
		size, err := request.Size(realURL, url)
		if err != nil {
			return nil, err
		}
		urlData := downloader.URL{
			URL:  realURL,
			Size: size,
			Ext:  "mp4",
		}
		streams[quality] = downloader.Stream{
			URLs:    []downloader.URL{urlData},
			Size:    size,
			Quality: fmt.Sprintf("%sP", quality),
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
