package tumblr

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/iawia002/annie/extractors/types"
	"github.com/iawia002/annie/parser"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type imageList struct {
	List []string `json:"@list"`
}

type tumblrImageList struct {
	Image imageList `json:"image"`
}

type tumblrImage struct {
	Image string `json:"image"`
}

func genURLData(url, referer string) (*types.Part, int64, error) {
	size, err := request.Size(url, referer)
	if err != nil {
		return nil, 0, err
	}
	_, ext, err := utils.GetNameAndExt(url)
	if err != nil {
		return nil, 0, err
	}
	return &types.Part{
		URL:  url,
		Size: size,
		Ext:  ext,
	}, size, nil
}

func tumblrImageDownload(url, html, title string) ([]*types.Data, error) {
	jsonStrings := utils.MatchOneOf(
		html, `<script type="application/ld\+json">\s*(.+?)</script>`,
	)
	if jsonStrings == nil || len(jsonStrings) < 2 {
		return nil, types.ErrURLParseFailed
	}
	jsonString := jsonStrings[1]

	var totalSize int64
	urls := make([]*types.Part, 0, 1)
	if strings.Contains(jsonString, `"image":{"@list"`) {
		// there are two data structures in the same field(image)
		var imageList tumblrImageList
		if err := json.Unmarshal([]byte(jsonString), &imageList); err != nil {
			return nil, err
		}
		for _, u := range imageList.Image.List {
			urlData, size, err := genURLData(u, url)
			if err != nil {
				return nil, err
			}
			totalSize += size
			urls = append(urls, urlData)
		}
	} else {
		var image tumblrImage
		if err := json.Unmarshal([]byte(jsonString), &image); err != nil {
			return nil, err
		}

		urlData, size, err := genURLData(image.Image, url)
		if err != nil {
			return nil, err
		}
		totalSize = size
		urls = append(urls, urlData)
	}
	streams := map[string]*types.Stream{
		"default": {
			Parts: urls,
			Size:  totalSize,
		},
	}

	return []*types.Data{
		{
			Site:    "Tumblr tumblr.com",
			Title:   title,
			Type:    types.DataTypeImage,
			Streams: streams,
			URL:     url,
		},
	}, nil
}

func tumblrVideoDownload(url, html, title string) ([]*types.Data, error) {
	videoURLs := utils.MatchOneOf(html, `<iframe src='(.+?)'`)
	if videoURLs == nil || len(videoURLs) < 2 {
		return nil, types.ErrURLParseFailed
	}
	videoURL := videoURLs[1]

	if !strings.Contains(videoURL, "tumblr.com/video") {
		return nil, errors.New("annie doesn't support this URL right now")
	}
	videoHTML, err := request.Get(videoURL, url, nil)
	if err != nil {
		return nil, err
	}

	realURLs := utils.MatchOneOf(videoHTML, `source src="(.+?)"`)
	if realURLs == nil || len(realURLs) < 2 {
		return nil, types.ErrURLParseFailed
	}
	realURL := realURLs[1]

	urlData, size, err := genURLData(realURL, url)
	if err != nil {
		return nil, err
	}
	streams := map[string]*types.Stream{
		"default": {
			Parts: []*types.Part{urlData},
			Size:  size,
		},
	}

	return []*types.Data{
		{
			Site:    "Tumblr tumblr.com",
			Title:   title,
			Type:    types.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
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
	if strings.Contains(html, "<iframe src=") {
		// Data
		return tumblrVideoDownload(url, html, title)
	}
	// Image
	return tumblrImageDownload(url, html, title)
}
