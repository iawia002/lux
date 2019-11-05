package pixivision

import (
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/parser"
	"github.com/iawia002/annie/request"
)

// Extract is the main function for extracting data
func Extract(url string) ([]downloader.Data, error) {
	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, err
	}
	title, urls, err := parser.GetImages(url, html, "am__work__illust  ", nil)
	if err != nil {
		return nil, err
	}
	streams := map[string]downloader.Stream{
		"default": {
			URLs: urls,
			Size: 0,
		},
	}

	return []downloader.Data{
		{
			Site:    "pixivision pixivision.net",
			Title:   title,
			Type:    "image",
			Streams: streams,
			URL:     url,
		},
	}, nil
}
