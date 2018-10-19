package instagram

import (
	"encoding/json"

	"github.com/iawia002/annie/downloader"
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

// Download main download function
func Download(url string) ([]downloader.Data, error) {
	html, err := request.Get(url, url, nil)
	if err != nil {
		return downloader.EmptyList, err
	}
	// get the title
	doc, err := parser.GetDoc(html)
	if err != nil {
		return downloader.EmptyList, err
	}
	title := parser.Title(doc)

	dataString := utils.MatchOneOf(html, `window\._sharedData\s*=\s*(.*);`)[1]
	var data instagram
	json.Unmarshal([]byte(dataString), &data)

	var realURL, dataType string
	var size int64
	streams := map[string]downloader.Stream{}

	if data.EntryData.PostPage[0].Graphql.ShortcodeMedia.VideoURL != "" {
		// Data
		dataType = "video"
		realURL = data.EntryData.PostPage[0].Graphql.ShortcodeMedia.VideoURL
		size, err = request.Size(realURL, url)
		if err != nil {
			return downloader.EmptyList, err
		}
		streams["default"] = downloader.Stream{
			URLs: []downloader.URL{
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
		dataType = "image"
		if data.EntryData.PostPage[0].Graphql.ShortcodeMedia.EdgeSidecar.Edges == nil {
			// Single
			realURL = data.EntryData.PostPage[0].Graphql.ShortcodeMedia.DisplayURL
			size, err = request.Size(realURL, url)
			if err != nil {
				return downloader.EmptyList, err
			}
			streams["default"] = downloader.Stream{
				URLs: []downloader.URL{
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
			var urls []downloader.URL
			for _, u := range data.EntryData.PostPage[0].Graphql.ShortcodeMedia.EdgeSidecar.Edges {
				realURL = u.Node.DisplayURL
				size, err = request.Size(realURL, url)
				if err != nil {
					return downloader.EmptyList, err
				}
				urlData := downloader.URL{
					URL:  realURL,
					Size: size,
					Ext:  "jpg",
				}
				urls = append(urls, urlData)
				totalSize += size
			}
			streams["default"] = downloader.Stream{
				URLs: urls,
				Size: totalSize,
			}
		}
	}

	return []downloader.Data{
		{
			Site:    "Instagram instagram.com",
			Title:   utils.FileName(title),
			Type:    dataType,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
