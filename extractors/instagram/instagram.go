package instagram

import (
	"encoding/json"
	netURL "net/url"
	"path"
	"strings"

	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/parser"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("instagram", New())
}

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

// New returns a instagram extractor.
func New() extractors.Extractor {
	return &extractor{}
}

func extractImageFromPage(html, url string) (map[string]*extractors.Stream, error) {
	_, realURLs, err := parser.GetImages(html, "EmbeddedMediaImage", nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	urls := make([]*extractors.Part, 0, len(realURLs))
	var totalSize int64
	for _, realURL := range realURLs {
		size, err := request.Size(realURL, url)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		urlData := &extractors.Part{
			URL:  realURL,
			Size: size,
			Ext:  "jpg",
		}
		urls = append(urls, urlData)
		totalSize += size
	}

	return map[string]*extractors.Stream{
		"default": {
			Parts: urls,
			Size:  totalSize,
		},
	}, nil
}

func extractFromData(dataString, url string) (map[string]*extractors.Stream, error) {
	var data instagram
	if err := json.Unmarshal([]byte(dataString), &data); err != nil {
		return nil, errors.WithStack(err)
	}

	urls := make([]*extractors.Part, 0, len(data.ShortcodeMedia.EdgeSidecar.Edges))
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
			return nil, errors.WithStack(err)
		}
		urlData := &extractors.Part{
			URL:  realURL,
			Size: size,
			Ext:  ext,
		}
		urls = append(urls, urlData)
		totalSize += size
	}

	return map[string]*extractors.Stream{
		"default": {
			Parts: urls,
			Size:  totalSize,
		},
	}, nil
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	// Instagram is forcing a login to access the page, so we use the embed page to bypass that.
	u, err := netURL.Parse(url)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	id := u.Path[strings.LastIndex(u.Path, "/")+1:]
	u.Path = path.Join(u.Path, "embed")

	html, err := request.Get(u.String(), url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	dataStrings := utils.MatchOneOf(html, `window\.__additionalDataLoaded\('graphql',(.*)\);`)
	if dataStrings == nil || len(dataStrings) < 2 {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}
	dataString := dataStrings[1]

	var streams map[string]*extractors.Stream
	if dataString == "" || dataString == "null" {
		streams, err = extractImageFromPage(html, url)
	} else {
		streams, err = extractFromData(dataString, url)
	}
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return []*extractors.Data{
		{
			Site:    "Instagram instagram.com",
			Title:   "Instagram " + id,
			Type:    extractors.DataTypeImage,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
