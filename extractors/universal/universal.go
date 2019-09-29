package universal

import (
	"fmt"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

// Extract is the main function for extracting data
func Extract(url string) ([]downloader.Data, error) {
	fmt.Println("\nannie doesn't support this URL right now, but it will try to download it directly")

	filename, ext, err := utils.GetNameAndExt(url)
	if err != nil {
		return nil, err
	}
	size, err := request.Size(url, url)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return []downloader.Data{
		{
			Site:    "Universal",
			Title:   filename,
			Type:    contentType,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
