package extractors

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

// Instagram download function
func Instagram(url string) downloader.VideoData {
	html := request.Get(url, url, nil)
	// get the title
	doc := parser.GetDoc(html)
	title := parser.Title(doc)

	dataString := utils.MatchOneOf(html, `window\._sharedData\s*=\s*(.*);`)[1]
	var data instagram
	json.Unmarshal([]byte(dataString), &data)

	var realURL, dataType string
	var size int64
	format := map[string]downloader.FormatData{}

	if data.EntryData.PostPage[0].Graphql.ShortcodeMedia.VideoURL != "" {
		// Video
		dataType = "video"
		realURL = data.EntryData.PostPage[0].Graphql.ShortcodeMedia.VideoURL
		size = request.Size(realURL, url)
		format["default"] = downloader.FormatData{
			URLs: []downloader.URLData{
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
			size = request.Size(realURL, url)
			format["default"] = downloader.FormatData{
				URLs: []downloader.URLData{
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
			var urls []downloader.URLData
			for _, u := range data.EntryData.PostPage[0].Graphql.ShortcodeMedia.EdgeSidecar.Edges {
				realURL = u.Node.DisplayURL
				size = request.Size(realURL, url)
				urlData := downloader.URLData{
					URL:  realURL,
					Size: size,
					Ext:  "jpg",
				}
				urls = append(urls, urlData)
				totalSize += size
			}
			format["default"] = downloader.FormatData{
				URLs: urls,
				Size: totalSize,
			}
		}
	}

	extractedData := downloader.VideoData{
		Site:    "Instagram instagram.com",
		Title:   utils.FileName(title),
		Type:    dataType,
		Formats: format,
	}
	extractedData.Download(url)
	return extractedData
}
