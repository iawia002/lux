package extractors

import (
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/parser"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

// Miaopai download function
func Miaopai(url string) downloader.VideoData {
	html := request.Get(url, url)
	doc := parser.GetDoc(html)
	title := parser.Title(doc)

	realURL := utils.MatchOneOf(html, `"videoSrc":"(.+?)"`)[1]
	size := request.Size(realURL, url)
	urlData := downloader.URLData{
		URL:  realURL,
		Size: size,
		Ext:  "mp4",
	}
	format := map[string]downloader.FormatData{
		"default": {
			URLs: []downloader.URLData{urlData},
			Size: size,
		},
	}
	extractedData := downloader.VideoData{
		Site:    "秒拍 miaopai.com",
		Title:   title,
		Type:    "video",
		Formats: format,
	}
	extractedData.Download(url)
	return extractedData
}
