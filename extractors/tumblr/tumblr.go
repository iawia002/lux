package tumblr

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/parser"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("tumblr", New())
}

type imageList struct {
	List []string `json:"@list"`
}

type tumblrImageList struct {
	Image imageList `json:"image"`
}

type tumblrImage struct {
	Image string `json:"image"`
}

func genURLData(url, referer string) (*extractors.Part, int64, error) {
	size, err := request.Size(url, referer)
	if err != nil {
		return nil, 0, err
	}
	_, ext, err := utils.GetNameAndExt(url)
	if err != nil {
		return nil, 0, err
	}
	return &extractors.Part{
		URL:  url,
		Size: size,
		Ext:  ext,
	}, size, nil
}

func tumblrImageDownload(url, html, title string) ([]*extractors.Data, error) {
	jsonStrings := utils.MatchOneOf(
		html, `<script type="application/ld\+json">\s*(.+?)</script>`,
	)
	if jsonStrings == nil || len(jsonStrings) < 2 {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}
	jsonString := jsonStrings[1]

	var totalSize int64
	urls := make([]*extractors.Part, 0, 1)
	if strings.Contains(jsonString, `"image":{"@list"`) {
		// there are two data structures in the same field(image)
		var imageList tumblrImageList
		if err := json.Unmarshal([]byte(jsonString), &imageList); err != nil {
			return nil, errors.WithStack(err)
		}
		for _, u := range imageList.Image.List {
			urlData, size, err := genURLData(u, url)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			totalSize += size
			urls = append(urls, urlData)
		}
	} else {
		var image tumblrImage
		if err := json.Unmarshal([]byte(jsonString), &image); err != nil {
			return nil, errors.WithStack(err)
		}

		urlData, size, err := genURLData(image.Image, url)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		totalSize = size
		urls = append(urls, urlData)
	}
	streams := map[string]*extractors.Stream{
		"default": {
			Parts: urls,
			Size:  totalSize,
		},
	}

	return []*extractors.Data{
		{
			Site:    "Tumblr tumblr.com",
			Title:   title,
			Type:    extractors.DataTypeImage,
			Streams: streams,
			URL:     url,
		},
	}, nil
}

func tumblrVideoDownload(url, html, title string) ([]*extractors.Data, error) {
	videoURLs := utils.MatchOneOf(html, `<iframe src='(.+?)'`)
	if videoURLs == nil || len(videoURLs) < 2 {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}
	videoURL := videoURLs[1]

	if !strings.Contains(videoURL, "tumblr.com/video") {
		return nil, errors.New("lux doesn't support this URL right now")
	}
	videoHTML, err := request.Get(videoURL, url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	realURLs := utils.MatchOneOf(videoHTML, `source src="(.+?)"`)
	if realURLs == nil || len(realURLs) < 2 {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}
	realURL := realURLs[1]

	urlData, size, err := genURLData(realURL, url)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	streams := map[string]*extractors.Stream{
		"default": {
			Parts: []*extractors.Part{urlData},
			Size:  size,
		},
	}

	return []*extractors.Data{
		{
			Site:    "Tumblr tumblr.com",
			Title:   title,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}

type extractor struct{}

// New returns a tumblr extractor.
func New() extractors.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// get the title
	doc, err := parser.GetDoc(html)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	title := parser.Title(doc)
	if strings.Contains(html, "<iframe src=") {
		// Data
		return tumblrVideoDownload(url, html, title)
	}
	// Image
	return tumblrImageDownload(url, html, title)
}
