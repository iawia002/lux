package tumblr

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/iawia002/annie/downloader"
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

func genURLData(url, referer string) (downloader.URL, int64, error) {
	size, err := request.Size(url, referer)
	if err != nil {
		return downloader.URL{}, 0, err
	}
	_, ext, err := utils.GetNameAndExt(url)
	if err != nil {
		return downloader.URL{}, 0, err
	}
	data := downloader.URL{
		URL:  url,
		Size: size,
		Ext:  ext,
	}
	return data, size, nil
}

func tumblrImageDownload(url, html, title string) ([]downloader.Data, error) {
	jsonString := utils.MatchOneOf(
		html, `<script type="application/ld\+json">\s*(.+?)</script>`,
	)[1]
	var totalSize int64
	var urls []downloader.URL
	if strings.Contains(jsonString, `"image":{"@list"`) {
		// there are two data structures in the same field(image)
		var imageList tumblrImageList
		json.Unmarshal([]byte(jsonString), &imageList)
		for _, u := range imageList.Image.List {
			urlData, size, err := genURLData(u, url)
			if err != nil {
				return downloader.EmptyList, err
			}
			totalSize += size
			urls = append(urls, urlData)
		}
	} else {
		var image tumblrImage
		json.Unmarshal([]byte(jsonString), &image)
		urlData, size, err := genURLData(image.Image, url)
		if err != nil {
			return downloader.EmptyList, err
		}
		totalSize = size
		urls = append(urls, urlData)
	}
	streams := map[string]downloader.Stream{
		"default": {
			URLs: urls,
			Size: totalSize,
		},
	}

	return []downloader.Data{
		{
			Site:    "Tumblr tumblr.com",
			Title:   title,
			Type:    "image",
			Streams: streams,
			URL:     url,
		},
	}, nil
}

func tumblrVideoDownload(url, html, title string) ([]downloader.Data, error) {
	videoURL := utils.MatchOneOf(html, `<iframe src='(.+?)'`)[1]
	if !strings.Contains(videoURL, "tumblr.com/video") {
		return downloader.EmptyList, errors.New("annie doesn't support this URL right now")
	}
	videoHTML, err := request.Get(videoURL, url, nil)
	if err != nil {
		return downloader.EmptyList, err
	}
	realURL := utils.MatchOneOf(videoHTML, `source src="(.+?)"`)[1]
	urlData, size, err := genURLData(realURL, url)
	if err != nil {
		return downloader.EmptyList, err
	}
	streams := map[string]downloader.Stream{
		"default": {
			URLs: []downloader.URL{urlData},
			Size: size,
		},
	}

	return []downloader.Data{
		{
			Site:    "Tumblr tumblr.com",
			Title:   title,
			Type:    "video",
			Streams: streams,
			URL:     url,
		},
	}, nil
}

// Download main download function
func Download(url string) ([]downloader.Data, error) {
	html, err := request.Get(url, url, nil)
	if err != nil {
		return downloader.EmptyList, err
	}
	// get the title
	doc, err := parser.GetDoc(html)
	if err != nil {
		return downloader.EmptyList, err
	}
	title := parser.Title(doc)
	if strings.Contains(html, "<iframe src=") {
		// Data
		return tumblrVideoDownload(url, html, title)
	}
	// Image
	return tumblrImageDownload(url, html, title)
}
