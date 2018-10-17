package miaopai

import (
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/parser"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

// Download main download function
func Download(url string) ([]downloader.Data, error) {
	html, err := request.Get(url, url, nil)
	if err != nil {
		return downloader.EmptyData, err
	}
	doc, err := parser.GetDoc(html)
	if err != nil {
		return downloader.EmptyData, err
	}
	title := parser.Title(doc)

	realURL := utils.MatchOneOf(html, `"videoSrc":"(.+?)"`)[1]
	size, err := request.Size(realURL, url)
	if err != nil {
		return downloader.EmptyData, err
	}
	urlData := downloader.URL{
		URL:  realURL,
		Size: size,
		Ext:  "mp4",
	}
	streams := map[string]downloader.Stream{
		"default": {
			URLs: []downloader.URL{urlData},
			Size: size,
		},
	}

	return []downloader.Data{
		{
			Site:    "秒拍 miaopai.com",
			Title:   title,
			Type:    "video",
			Streams: streams,
		},
	}, nil
}
