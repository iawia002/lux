package extractors

import (
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/parser"
	"github.com/iawia002/annie/request"
)

// Pixivision download function
func Pixivision(url string) (downloader.VideoData, error) {
	html, err := request.Get(url, url, nil)
	if err != nil {
		return downloader.VideoData{}, err
	}
	title, urls, err := parser.GetImages(url, html, "am__work__illust  ", nil)
	if err != nil {
		return downloader.VideoData{}, err
	}
	format := map[string]downloader.FormatData{
		"default": {
			URLs: urls,
			Size: 0,
		},
	}
	extractedData := downloader.VideoData{
		Site:    "pixivision pixivision.net",
		Title:   title,
		Type:    "image",
		Formats: format,
	}
	err = extractedData.Download(url)
	return extractedData, nil
}
