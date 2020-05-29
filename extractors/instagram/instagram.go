package instagram

import (
	"encoding/json"
	netURL "net/url"
	"path"
	"strings"

	"github.com/iawia002/annie/extractors/types"
	"github.com/iawia002/annie/parser"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type instagram struct {
	ShortcodeMedia struct {
		EdgeSidecar struct {
			Edges []struct {
				Node struct {
					DisplayURL string `json:"display_url"`
					IsVideo    bool   `json:"is_video"`
					VideoURL   string `json:"video_url"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"edge_sidecar_to_children"`
	} `json:"shortcode_media"`
}

type extractor struct{}

// New returns a youtube extractor.
func New() types.Extractor {
	return &extractor{}
}

func extractImageFromPage(html, url string) (map[string]*types.Stream, error) {
	_, realURLs, err := parser.GetImages(html, "EmbeddedMediaImage", nil)
	if err != nil {
		return nil, err
	}

	urls := make([]*types.Part, 0, len(realURLs))
	var totalSize int64
	for _, realURL := range realURLs {
		size, err := request.Size(realURL, url)
		if err != nil {
			return nil, err
		}
		urlData := &types.Part{
			URL:  realURL,
			Size: size,
			Ext:  "jpg",
		}
		urls = append(urls, urlData)
		totalSize += size
	}

	return map[string]*types.Stream{
		"default": {
			Parts: urls,
			Size:  totalSize,
		},
	}, nil
}

func extractFromData(dataString, url string) (map[string]*types.Stream, error) {
	var data instagram
	if err := json.Unmarshal([]byte(dataString), &data); err != nil {
		return nil, err
	}

	urls := make([]*types.Part, 0, len(data.ShortcodeMedia.EdgeSidecar.Edges))
	var totalSize int64
	for _, u := range data.ShortcodeMedia.EdgeSidecar.Edges {
		// Image
		realURL := u.Node.DisplayURL
		ext := "jpg"
		if u.Node.IsVideo {
			// Video
			realURL = u.Node.VideoURL
			ext = "mp4"
		}

		size, err := request.Size(realURL, url)
		if err != nil {
			return nil, err
		}
		urlData := &types.Part{
			URL:  realURL,
			Size: size,
			Ext:  ext,
		}
		urls = append(urls, urlData)
		totalSize += size
	}

	return map[string]*types.Stream{
		"default": {
			Parts: urls,
			Size:  totalSize,
		},
	}, nil
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option types.Options) ([]*types.Data, error) {
	// Instagram is forcing a login to access the page, so we use the embed page to bypass that.
	u, err := netURL.Parse(url)
	if err != nil {
		return nil, err
	}
	id := u.Path[strings.LastIndex(u.Path, "/")+1:]
	u.Path = path.Join(u.Path, "embed")

	html, err := request.Get(u.String(), url, nil)
	if err != nil {
		return nil, err
	}
	dataStrings := utils.MatchOneOf(html, `window\.__additionalDataLoaded\('graphql',(.*)\);`)
	if dataStrings == nil || len(dataStrings) < 2 {
		return nil, types.ErrURLParseFailed
	}
	dataString := dataStrings[1]

	var streams map[string]*types.Stream
	if dataString == "" || dataString == "null" {
		streams, err = extractImageFromPage(html, url)
	} else {
		streams, err = extractFromData(dataString, url)
	}
	if err != nil {
		return nil, err
	}

	return []*types.Data{
		{
			Site:    "Instagram instagram.com",
			Title:   "Instagram " + id,
			Type:    types.DataTypeImage,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
