package universal

import (
	"fmt"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

// Download main download function
func Download(url string) ([]downloader.Data, error) {
	fmt.Println()
	fmt.Println("annie doesn't support this URL right now, but it will try to download it directly")

	filename, ext, err := utils.GetNameAndExt(url)
	if err != nil {
		return downloader.EmptyList, err
	}
	size, err := request.Size(url, url)
	if err != nil {
		return downloader.EmptyList, err
	}
	urlData := downloader.URL{
		URL:  url,
		Size: size,
		Ext:  ext,
	}
	streams := map[string]downloader.Stream{
		"default": {
			URLs: []downloader.URL{urlData},
			Size: size,
		},
	}
	contentType, err := request.ContentType(url, url)
	if err != nil {
		return downloader.EmptyList, err
	}

	return []downloader.Data{
		{
			Site:    "Universal",
			Title:   utils.FileName(filename),
			Type:    contentType,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
