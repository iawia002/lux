package instagram

import (
	"encoding/json"

	"github.com/iawia002/annie/extractors/types"
	"github.com/iawia002/annie/parser"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type instagram struct {
	EntryData struct {
		PostPage []struct {
			Graphql struct {
				ShortcodeMedia struct {
					DisplayURL  string `json:"display_url"`
					VideoURL    string `json:"video_url"`
					EdgeSidecar struct {
						Edges []struct {
							Node struct {
								DisplayURL string `json:"display_url"`
							} `json:"node"`
						} `json:"edges"`
					} `json:"edge_sidecar_to_children"`
				} `json:"shortcode_media"`
			} `json:"graphql"`
		} `json:"PostPage"`
	} `json:"entry_data"`
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
	// get the title
	doc, err := parser.GetDoc(html)
	if err != nil {
		return nil, err
	}
	title := parser.Title(doc)

	dataStrings := utils.MatchOneOf(html, `window\._sharedData\s*=\s*(.*);`)
	if dataStrings == nil || len(dataStrings) < 2 {
		return nil, types.ErrURLParseFailed
	}
	dataString := dataStrings[1]

	var data instagram
	if err = json.Unmarshal([]byte(dataString), &data); err != nil {
		return nil, types.ErrURLParseFailed
	}

	var (
		realURL  string
		dataType types.DataType
		size     int64
	)
	streams := make(map[string]*types.Stream)

	if data.EntryData.PostPage[0].Graphql.ShortcodeMedia.VideoURL != "" {
		// Video
		dataType = types.DataTypeVideo
		realURL = data.EntryData.PostPage[0].Graphql.ShortcodeMedia.VideoURL
		size, err = request.Size(realURL, url)
		if err != nil {
			return nil, err
		}
		streams["default"] = &types.Stream{
			Parts: []*types.Part{
				{
					URL:  realURL,
					Size: size,
					Ext:  "mp4",
				},
			},
			Size: size,
		}
	} else {
		// Image
		dataType = types.DataTypeImage
		if data.EntryData.PostPage[0].Graphql.ShortcodeMedia.EdgeSidecar.Edges == nil {
			// Single
			realURL = data.EntryData.PostPage[0].Graphql.ShortcodeMedia.DisplayURL
			size, err = request.Size(realURL, url)
			if err != nil {
				return nil, err
			}
			streams["default"] = &types.Stream{
				Parts: []*types.Part{
					{
						URL:  realURL,
						Size: size,
						Ext:  "jpg",
					},
				},
				Size: size,
			}
		} else {
			// Album
			var totalSize int64
			var urls []*types.Part
			for _, u := range data.EntryData.PostPage[0].Graphql.ShortcodeMedia.EdgeSidecar.Edges {
				realURL = u.Node.DisplayURL
				size, err = request.Size(realURL, url)
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
			streams["default"] = &types.Stream{
				Parts: urls,
				Size:  totalSize,
			}
		}
	}

	return []*types.Data{
		{
			Site:    "Instagram instagram.com",
			Title:   title,
			Type:    dataType,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
