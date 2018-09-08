package extractors

import (
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/parser"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

// Miaopai download function
func Miaopai(url string) (downloader.VideoData, error) {
	html, err := request.Get(url, url, nil)
	if err != nil {
		return downloader.VideoData{}, err
	}
	doc, err := parser.GetDoc(html)
	if err != nil {
		return downloader.VideoData{}, err
	}
	title := parser.Title(doc)

	realURL := utils.MatchOneOf(html, `"videoSrc":"(.+?)"`)[1]
	size, err := request.Size(realURL, url)
	if err != nil {
		return downloader.VideoData{}, err
	}
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
	err = extractedData.Download(url)
	if err != nil {
		return downloader.VideoData{}, err
	}
	return extractedData, nil
}
