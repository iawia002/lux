package extractors

import (
	"encoding/json"
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

func genURLData(url, referer string) (downloader.URLData, int64) {
	size := request.Size(url, referer)
	_, ext := utils.GetNameAndExt(url)
	data := downloader.URLData{
		URL:  url,
		Size: size,
		Ext:  ext,
	}
	return data, size
}

// Tumblr download function
func Tumblr(url string) downloader.VideoData {
	html := request.Get(url)
	doc := parser.GetDoc(html)
	title := strings.TrimSpace(doc.Find("title").Text())
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
			urlData, size := genURLData(u, url)
			totalSize += size
			urls = append(urls, urlData)
		}
	} else {
		var image tumblrImage
		json.Unmarshal([]byte(jsonString), &image)
		urlData, size := genURLData(image.Image, url)
		totalSize = size
		urls = append(urls, urlData)
	}

	data := downloader.VideoData{
		Site:  "Tumblr tumblr.com",
		Title: title,
		Type:  "image",
		URLs:  urls,
		Size:  totalSize,
	}
	data.Download(url)
	return data
}
