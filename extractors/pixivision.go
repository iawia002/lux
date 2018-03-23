package extractors

import (
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/parser"
	"github.com/iawia002/annie/request"
)

// Pixivision download function
func Pixivision(url string) downloader.VideoData {
	html := request.Get(url)
	title, urls := parser.GetImages(url, html, "am__work__illust  ", nil)
	format := map[string]downloader.FormatData{
		"default": downloader.FormatData{
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
	extractedData.Download(url)
	return extractedData
}
