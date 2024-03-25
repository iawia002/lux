package instagram

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	netURL "net/url"
	"regexp"
	"strings"
	"time"

	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/gocolly/colly/v2"
	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

var client *http.Client

func init() {
	extractors.Register("instagram", New())
	client = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 5 * time.Second,
		},
	}
}

// sliderItemNode contains information about the Instagram post
type sliderItemNode struct {
	DisplayURL string `json:"display_url"` // URL of the Media (resolution is dynamic)

	IsVideo  bool   `json:"is_video"`  // Is type of the Media equals to video
	VideoURL string `json:"video_url"` // Direct URL to the Video
}

func (s sliderItemNode) extractMediaURL() string {
	if s.IsVideo {
		return s.VideoURL
	}

	return s.DisplayURL
}

type instagramPayload struct {
	Media struct {
		ID          string `json:"id"` // Unique ID of the Media
		SliderItems struct {
			Edges []struct {
				Node sliderItemNode `json:"node"`
			} `json:"edges"`
		} `json:"edge_sidecar_to_children"` // Children of the Media
	} `json:"shortcode_media"` // Media
}

func (s instagramPayload) isEmpty() bool {
	return s.Media.ID == ""
}

func getPostWithCode(code string) ([]string, error) {
	URL := fmt.Sprintf("https://www.instagram.com/p/%v/embed/captioned/", code)

	var embeddedMediaImage string
	var embedResponse = instagramPayload{}
	collector := colly.NewCollector()
	collector.SetClient(client)
	var collectorErr error

	collector.OnHTML("img.EmbeddedMediaImage", func(e *colly.HTMLElement) {
		embeddedMediaImage = e.Attr("src")
	})

	collector.OnHTML("script", func(e *colly.HTMLElement) {
		r := regexp.MustCompile(`\\\"gql_data\\\":([\s\S]*)\}\"\}\]\]\,\[\"NavigationMetrics`)
		match := r.FindStringSubmatch(e.Text)

		if len(match) < 2 {
			return
		}

		s := strings.ReplaceAll(match[1], `\"`, `"`)
		s = strings.ReplaceAll(s, `\\/`, `/`)
		s = strings.ReplaceAll(s, `\\`, `\`)

		err := json.Unmarshal([]byte(s), &embedResponse)
		if err != nil {
			collectorErr = err
		}
	})

	collector.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", browser.Chrome())
	})

	if err := collector.Visit(URL); err != nil {
		return nil, fmt.Errorf("failed to send HTTP request to the Instagram: %v", err)
	}

	if collectorErr != nil {
		return nil, fmt.Errorf("failed to parse the Instagram response: %v", collectorErr)
	}

	// If the method one which is JSON parsing didn't fail
	if !embedResponse.isEmpty() {
		result := make([]string, 0, len(embedResponse.Media.SliderItems.Edges))
		for _, item := range embedResponse.Media.SliderItems.Edges {
			result = append(result, item.Node.extractMediaURL())
		}

		return result, nil
	}

	if embeddedMediaImage != "" {
		return []string{embeddedMediaImage}, nil
	}

	// If every two methods have failed, then return an error
	return nil, errors.New("failed to fetch the post, the page might be \"private\", or the link is completely wrong")
}

func extractShortCodeFromLink(link string) (string, error) {
	values := regexp.MustCompile(`(p|tv|reel|reels\/videos)\/([A-Za-z0-9-_]+)`).FindStringSubmatch(link)
	if len(values) != 3 {
		return "", errors.New("couldn't extract the media short code from the link")
	}

	return values[2], nil
}

type extractor struct{}

// New returns a instagram extractor.
func New() extractors.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	u, err := netURL.Parse(url)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	shortCode, err := extractShortCodeFromLink(u.String())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	urls, err := getPostWithCode(shortCode)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var totalSize int64
	var parts []*extractors.Part

	for _, u := range urls {
		_, ext, err := utils.GetNameAndExt(u)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		fileSize, err := request.Size(u, url)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		part := &extractors.Part{
			URL:  u,
			Size: fileSize,
			Ext:  ext,
		}
		parts = append(parts, part)
	}

	for _, part := range parts {
		totalSize += part.Size
	}

	streams := map[string]*extractors.Stream{
		"default": {
			Parts: parts,
			Size:  totalSize,
		},
	}

	return []*extractors.Data{
		{
			Site:    "Instagram instagram.com",
			Title:   "Instagram " + shortCode,
			Type:    extractors.DataTypeImage,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
