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

func genURLData(url, referer string) (downloader.URLData, int64, error) {
	size, err := request.Size(url, referer)
	if err != nil {
		return downloader.URLData{}, 0, err
	}
	_, ext, err := utils.GetNameAndExt(url)
	if err != nil {
		return downloader.URLData{}, 0, err
	}
	data := downloader.URLData{
		URL:  url,
		Size: size,
		Ext:  ext,
	}
	return data, size, nil
}

func tumblrImageDownload(url, html, title string) ([]downloader.VideoData, error) {
	jsonString := utils.MatchOneOf(
		html, `<script type="application/ld\+json">\s*(.+?)</script>`,
	)[1]
	var totalSize int64
	var urls []downloader.URLData
	if strings.Contains(jsonString, `"image":{"@list"`) {
		// there are two data structures in the same field(image)
		var imageList tumblrImageList
		json.Unmarshal([]byte(jsonString), &imageList)
		for _, u := range imageList.Image.List {
			urlData, size, err := genURLData(u, url)
			if err != nil {
				return downloader.EmptyData, err
			}
			totalSize += size
			urls = append(urls, urlData)
		}
	} else {
		var image tumblrImage
		json.Unmarshal([]byte(jsonString), &image)
		urlData, size, err := genURLData(image.Image, url)
		if err != nil {
			return downloader.EmptyData, err
		}
		totalSize = size
		urls = append(urls, urlData)
	}
	format := map[string]downloader.FormatData{
		"default": {
			URLs: urls,
			Size: totalSize,
		},
	}

	return []downloader.VideoData{
		{
			Site:    "Tumblr tumblr.com",
			Title:   title,
			Type:    "image",
			Formats: format,
		},
	}, nil
}

func tumblrVideoDownload(url, html, title string) ([]downloader.VideoData, error) {
	videoURL := utils.MatchOneOf(html, `<iframe src='(.+?)'`)[1]
	if !strings.Contains(videoURL, "tumblr.com/video") {
		return downloader.EmptyData, errors.New("annie doesn't support this URL right now")
	}
	videoHTML, err := request.Get(videoURL, url, nil)
	if err != nil {
		return downloader.EmptyData, err
	}
	realURL := utils.MatchOneOf(videoHTML, `source src="(.+?)"`)[1]
	urlData, size, err := genURLData(realURL, url)
	if err != nil {
		return downloader.EmptyData, err
	}
	format := map[string]downloader.FormatData{
		"default": {
			URLs: []downloader.URLData{urlData},
			Size: size,
		},
	}

	return []downloader.VideoData{
		{
			Site:    "Tumblr tumblr.com",
			Title:   title,
			Type:    "video",
			Formats: format,
		},
	}, nil
}

// Download main download function
func Download(url string) ([]downloader.VideoData, error) {
	html, err := request.Get(url, url, nil)
	if err != nil {
		return downloader.EmptyData, err
	}
	// get the title
	doc, err := parser.GetDoc(html)
	if err != nil {
		return downloader.EmptyData, err
	}
	title := parser.Title(doc)
	if strings.Contains(html, "<iframe src=") {
		// Video
		return tumblrVideoDownload(url, html, title)
	}
	// Image
	return tumblrImageDownload(url, html, title)
}
