package extractors

import (
	"fmt"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

// Universal download function
func Universal(url string) downloader.VideoData {
	fmt.Println()
	fmt.Println("annie doesn't support this URL right now, but it will try to download it directly")

	filename, ext := utils.GetNameAndExt(url)
	size := request.Size(url, url)
	urlData := downloader.URLData{
		URL:  url,
		Size: size,
		Ext:  ext,
	}
	format := map[string]downloader.FormatData{
		"default": downloader.FormatData{
			URLs: []downloader.URLData{urlData},
			Size: size,
		},
	}
	extractedData := downloader.VideoData{
		Site:    "Universal",
		Title:   utils.FileName(filename),
		Type:    request.ContentType(url, url),
		Formats: format,
	}
	extractedData.Download(url)
	return extractedData
}
