package bcy

import (
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/parser"
	"github.com/iawia002/annie/request"
)

// Download main download function
func Download(url string) ([]downloader.Data, error) {
	var err error
	html, err := request.Get(url, url, nil)
	if err != nil {
		return downloader.EmptyList, err
	}
	title, urls, err := parser.GetImages(
		url, html, "detail_std detail_clickable", func(u string) string {
			// https://img9.bcyimg.com/drawer/15294/post/1799t/1f5a87801a0711e898b12b640777720f.jpg/w650
			return u[:len(u)-5]
		},
	)
	if err != nil {
		return downloader.EmptyList, err
	}
	streams := map[string]downloader.Stream{
		"default": {
			URLs: urls,
			Size: 0,
		},
	}
	return []downloader.Data{
		{
			Site:    "半次元 bcy.net",
			Title:   title,
			Type:    "image",
			Streams: streams,
			URL:     url,
		},
	}, nil
}
