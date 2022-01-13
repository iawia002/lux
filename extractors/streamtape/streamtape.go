package streamtape

import (
	"regexp"
	"strings"

	"github.com/iawia002/lux/extractors/types"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

const prefix = "document.getElementById('robotlink').innerHTML = '"

var pattern = regexp.MustCompile(`\((.*?)\)`)

type extractor struct{}

// New returns a StreamTape extractor
func New() types.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, _ types.Options) ([]*types.Data, error) {
	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, err
	}

	var u string
	for _, line := range strings.Split(html, "\n") {
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		start := line[len(prefix):]

		domain := "https:" + start[:strings.Index(start, "'")]
		paramsMatches := pattern.FindAllStringSubmatch(start, -1)
		if len(paramsMatches) < 2 {
			return nil, types.ErrURLParseFailed
		}
		params := paramsMatches[0][1]
		params = params[1 : len(params)-1]

		u = domain + params[3:] + "&stream=1"
		break
	}
	if u == "" {
		return nil, types.ErrURLParseFailed
	}

	// get title
	var title = "StreamTape Video"
	titleMatch := utils.MatchOneOf(html,
		`\<meta name="og:title" content="(.*)"\>`)
	if len(titleMatch) >= 2 {
		title = titleMatch[1]
	}

	size, err := request.Size(u, url)
	if err != nil {
		return nil, err
	}

	urlData := &types.Part{
		URL:  u,
		Size: size,
		Ext:  "mp4",
	}

	streams := make(map[string]*types.Stream)
	streams["default"] = &types.Stream{
		Parts: []*types.Part{urlData},
		Size:  size,
	}

	return []*types.Data{
		{
			URL:     u,
			Site:    "StreamTape streamtape.com",
			Title:   title,
			Type:    types.DataTypeVideo,
			Streams: streams,
		},
	}, nil
}
