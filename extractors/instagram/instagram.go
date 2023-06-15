package instagram

import (
	"encoding/json"
	netURL "net/url"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/net/html"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
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

type instagramPayload struct {
    ArticleBody string `json:"articleBody"`
    Author struct {
        Image string `json:"image"`
        Name string `json:"name"`
        AlternativeName string `json:"alternativeName"`
        Url string `json:"url"`
    }`json:"author"`
    Videos []struct {
        UploadData string `json:"string"`
        Description string `json:"description"`
        Name string `json:"name"`
        Caption string `json:"caption"`
        Height string `json:"height"`
        Width string `json:"width"`
        ContentURL string `json:"contentUrl"`
        ThumbnailURL string `json:"thumbnailUrl"`
    } `json:"video"`
    Images []struct {
        Caption string `json:"caption"`
        Height string `json:"height"`
        Width string `json:"width"`
        Url string `json:"url"`
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

    var parts []*extractors.Part
    if len(payload.Videos) > 0 {
        for _, it := range payload.Videos {
            ext := ""
            part := &extractors.Part{
                URL:  it.ContentURL,
                Size: 0,
                Ext:  ext,
            }
            parts = append(parts, part)
        }
    } else if len(payload.Images) > 0 {
        for _, it := range payload.Videos {
            ext := ""
            part := &extractors.Part{
                URL:  it.ContentURL,
                Size: 0,
                Ext:  ext,
            }
            parts = append(parts, part)
        }
    } else {
        return nil, errors.WithStack(extractors.ErrBodyParseFailed)
    }
            
    streams :=  map[string]*extractors.Stream{
        "default": {
            Parts: parts,
            Size:  0,
        },
    };
    
    id := u.Path[strings.LastIndex(u.Path, "/") + 1:]

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
