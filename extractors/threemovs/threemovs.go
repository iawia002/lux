package threemovs

import (
	"fmt"

	"github.com/iawia002/annie/extractors/types"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type extractor struct{}

// New returns a 3movs extractor.
func New() types.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option types.Options) ([]*types.Data, error) {
	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, err
	}

	var (
		title  string
		url360 string
		url480 string

		streams       = make(map[string]*types.Stream, 2)
		finalRealURLs = make(map[string]string, 2)

		matchTitle = utils.MatchOneOf(html, `h1>(.*)<\/`)
		match360   = utils.MatchOneOf(html, `download_text_link_2" href="([^"]+)`)
		match480   = utils.MatchOneOf(html, `function\/0\/(.*?)'`)
		matchRnd   = utils.MatchOneOf(html, `rnd: '(\d+)`)
	)

	if len(matchTitle) < 2 {
		return nil, types.ErrURLParseFailed
	}

	if len(match360) < 2 {
		return nil, types.ErrURLParseFailed
	}

	if len(match480) < 2 {
		return nil, types.ErrURLParseFailed
	}

	if len(matchRnd) < 2 {
		return nil, types.ErrURLParseFailed
	}

	title = matchTitle[1]
	url360 = match360[1]
	url480 = fmt.Sprintf("%s?rnd=%s", match480[1], matchRnd[1])

	finalRealURLs["360"] = url360
	finalRealURLs["480"] = url480

	for key := range finalRealURLs {
		realURL := finalRealURLs[key]
		size, err := request.Size(realURL, url)
		if err != nil {
			return nil, err
		}

		urlData := &types.Part{
			URL:  realURL,
			Size: size,
			Ext:  "mp4",
		}

		streams[key] = &types.Stream{
			Parts:   []*types.Part{urlData},
			Size:    size,
			Quality: fmt.Sprintf("%sp", key),
		}
	}

	return []*types.Data{
		{
			Site:    "3movs 3movs.com",
			Title:   title,
			Type:    types.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
