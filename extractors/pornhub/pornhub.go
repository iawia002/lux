package pornhub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/iawia002/annie/extractors/types"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"

	"github.com/robertkrimen/otto"
)

type pornhubData struct {
	Format   string          `json:"format"`
	Quality  json.RawMessage `json:"quality"`
	VideoURL string          `json:"videoUrl"`
}

type extractor struct{}

// New returns a youtube extractor.
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

	realURLs := utils.MatchOneOf(html, `"mediaDefinitions":(.+?),"isVertical"`)
	if realURLs == nil || len(realURLs) < 2 {
		return nil, types.ErrURLParseFailed
	}

	var pornhubs []pornhubData
	if err = json.Unmarshal([]byte(realURLs[1]), &pornhubs); err != nil {
		return nil, err
	}

	scripts := utils.MatchAll(html, `<script type="text/javascript">(?s)(.+?)</script>`)
	var flashvarJS string
	for _, s := range scripts {
		for _, js := range s[1:] {
			if strings.Contains(js, "flashvars_") {
				flashvarJS = js
				break
			}
		}
	}
	vm := otto.New()
	vm.Run(flashvarJS) // nolint
	var urls []string
	for i := 0; i < len(pornhubs); i++ {
		if value, err := vm.Get(fmt.Sprintf("media_%d", i)); err == nil {
			urls = append(urls, value.String())
		}
	}

	streams := make(map[string]*types.Stream, len(pornhubs))
	for i, data := range pornhubs {
		if data.Format != "mp4" {
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

		realURL := urls[i]
		if len(realURL) == 0 {
			continue
		}
		size, err := request.Size(realURL, url)
		if err != nil {
			return nil, err
		}
		urlData := &types.Part{
			URL:  realURL,
			Size: size,
			Ext:  "mp4",
		}
		streams[quality] = &types.Stream{
			Parts:   []*types.Part{urlData},
			Size:    size,
			Quality: fmt.Sprintf("%sP", quality),
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
