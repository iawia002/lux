package instagram

import (
	"encoding/json"
	netURL "net/url"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/net/html"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("instagram", New())
}

type instagramPayload struct {
	ArticleBody string `json:"articleBody"`
	Author      struct {
		Image           string `json:"image"`
		Name            string `json:"name"`
		AlternativeName string `json:"alternativeName"`
		Url             string `json:"url"`
	} `json:"author"`
	Videos []struct {
		UploadData   string `json:"string"`
		Description  string `json:"description"`
		Name         string `json:"name"`
		Caption      string `json:"caption"`
		Height       string `json:"height"`
		Width        string `json:"width"`
		ContentURL   string `json:"contentUrl"`
		ThumbnailURL string `json:"thumbnailUrl"`
	} `json:"video"`
	Images []struct {
		Caption string `json:"caption"`
		Height  string `json:"height"`
		Width   string `json:"width"`
		URL     string `json:"url"`
	} `json:"image"`
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

	htmlResp, err := request.Get(u.String(), url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	reader := strings.NewReader(htmlResp)
	htmlRoot, err := html.Parse(reader)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	sNode, err := dfsFindScript(htmlRoot)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var payload instagramPayload
	if err = json.Unmarshal([]byte(sNode.Data), &payload); err != nil {
		return nil, errors.WithStack(err)
	}

	var totalSize int64
	var parts []*extractors.Part
	if len(payload.Videos) > 0 {
		videoParts, err := createPartVideos(&payload, url)
		if err != nil {
			return nil, errors.WithStack(extractors.ErrBodyParseFailed)
		}

		parts = append(parts, videoParts...)
	}
	if len(payload.Images) > 0 {
		imageParts, err := createPartImages(&payload, url)
		if err != nil {
			return nil, errors.WithStack(extractors.ErrBodyParseFailed)
		}

		parts = append(parts, imageParts...)
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

	id := u.Path[strings.LastIndex(u.Path, "/")+1:]

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

func dfsFindScript(n *html.Node) (*html.Node, error) {
	if n.Type == html.ElementNode && n.Data == "script" {
		for _, attr := range n.Attr {
			if attr.Key == "type" && attr.Val == "application/ld+json" {
				return n.FirstChild, nil
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if ret, err := dfsFindScript(c); err == nil {
			return ret, nil
		}
	}

	return nil, errors.WithStack(extractors.ErrBodyParseFailed)
}

func createPartVideos(payload *instagramPayload, ref string) (parts []*extractors.Part, err error) {
	for _, it := range payload.Videos {
		_, ext, err := utils.GetNameAndExt(it.ContentURL)
		if err != nil {
			return parts, errors.WithStack(err)
		}
		filesize, err := request.Size(it.ContentURL, ref)
		if err != nil {
			return parts, errors.WithStack(err)
		}

		part := &extractors.Part{
			URL:  it.ContentURL,
			Size: filesize,
			Ext:  ext,
		}
		parts = append(parts, part)
	}

	return parts, err
}

func createPartImages(payload *instagramPayload, ref string) (parts []*extractors.Part, err error) {
	for _, it := range payload.Images {
		_, ext, err := utils.GetNameAndExt(it.URL)
		if err != nil {
			return parts, errors.WithStack(err)
		}
		filesize, err := request.Size(it.URL, ref)
		if err != nil {
			return parts, errors.WithStack(err)
		}

		part := &extractors.Part{
			URL:  it.URL,
			Size: filesize,
			Ext:  ext,
		}
		parts = append(parts, part)
	}

	return parts, err
}
