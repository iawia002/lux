package universal

import (
	"fmt"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

// Download main download function
func Download(url string) ([]downloader.VideoData, error) {
	fmt.Println()
	fmt.Println("annie doesn't support this URL right now, but it will try to download it directly")

	filename, ext, err := utils.GetNameAndExt(url)
	if err != nil {
		return downloader.EmptyData, err
	}
	size, err := request.Size(url, url)
	if err != nil {
		return downloader.EmptyData, err
	}
	urlData := downloader.URLData{
		URL:  url,
		Size: size,
		Ext:  ext,
	}
	format := map[string]downloader.FormatData{
		"default": {
			URLs: []downloader.URLData{urlData},
			Size: size,
		},
	}
	contentType, err := request.ContentType(url, url)
	if err != nil {
		return downloader.EmptyData, err
	}

	return []downloader.VideoData{
		{
			Site:    "Universal",
			Title:   utils.FileName(filename),
			Type:    contentType,
			Formats: format,
		},
	}, nil
}
