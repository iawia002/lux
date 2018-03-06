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
	fmt.Println("annie doesn't support this URL by now, but it will try to download it directly")

	filename, ext := utils.GetNameAndExt(url)
	size := request.Size(url, url)
	urlData := downloader.URLData{
		URL:  url,
		Size: size,
		Ext:  ext,
	}
	data := downloader.VideoData{
		Site:  "Universal",
		Title: utils.FileName(filename),
		Type:  request.ContentType(url, url),
		URLs:  []downloader.URLData{urlData},
		Size:  size,
	}
	data.Download(url)
	return data
}
