package pornhub

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/iawia002/lux/extractors/types"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

type pornhubData struct {
	Quality  json.RawMessage `json:"text"`
	VideoURL string          `json:"url"`
}

type extractor struct{}

// New returns a pornhub extractor.
func New() types.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option types.Options) ([]*types.Data, error) {
	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, err
	}

	var title string
	desc := utils.MatchOneOf(html, `<span class="inlineFree">(.+?)</span>`)
	if len(desc) > 1 {
		title = desc[1]
	} else {
		title = "pornhub video"
	}

	realURLs := utils.MatchOneOf(html, `qualityItems_\d+\s=\s([^;]+)`)
	if realURLs == nil || len(realURLs) < 2 {
		return nil, types.ErrURLParseFailed
	}

	var pornhubs []pornhubData
	if err = json.Unmarshal([]byte(realURLs[1]), &pornhubs); err != nil {
		return nil, err
	}

	streams := make(map[string]*types.Stream, len(pornhubs))
	for _, data := range pornhubs {
		if !strings.Contains(data.VideoURL, "mp4") {
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
		quality := strings.ReplaceAll(string(data.Quality), "\"", "")

		size, err := request.Size(data.VideoURL, data.VideoURL)
		if err != nil {
			return nil, err
		}
		urlData := &types.Part{
			URL:  data.VideoURL,
			Size: size,
			Ext:  "mp4",
		}
		streams[quality] = &types.Stream{
			Parts:   []*types.Part{urlData},
			Size:    size,
			Quality: quality,
		}
	}

	return []*types.Data{
		{
			Site:    "Pornhub pornhub.com",
			Title:   title,
			Type:    types.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
