package extractors

import (
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/parser"
	"github.com/iawia002/annie/request"
)

// Bcy download function
func Bcy(url string) downloader.VideoData {
	html := request.Get(url)
	title, urls := parser.GetImages(
		url, html, "detail_std detail_clickable", func(u string) string {
			// https://img9.bcyimg.com/drawer/15294/post/1799t/1f5a87801a0711e898b12b640777720f.jpg/w650
			return u[:len(u)-5]
		},
	)
	format := map[string]downloader.FormatData{
		"default": downloader.FormatData{
			URLs: urls,
			Size: 0,
		},
	}
	extractedData := downloader.VideoData{
		Site:    "半次元 bcy.net",
		Title:   title,
		Type:    "image",
		Formats: format,
	}
	extractedData.Download(url)
	return extractedData
}
