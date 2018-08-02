package extractors

import (
	"fmt"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

// Facebook download function
func Facebook(url string) downloader.VideoData {
	html := request.Get(url, url, nil)
	title := utils.MatchOneOf(html, `<title id="pageTitle">(.+)</title>`)[1]

	format := map[string]downloader.FormatData{}
	for _, quality := range []string{"sd", "hd"} {
		u := utils.MatchOneOf(
			html, fmt.Sprintf(`%s_src:"(.+?)"`, quality),
		)[1]
		size := request.Size(u, url)
		urlData := downloader.URLData{
			URL:  u,
			Size: size,
			Ext:  "mp4",
		}
		format[quality] = downloader.FormatData{
			URLs:    []downloader.URLData{urlData},
			Size:    size,
			Quality: quality,
		}
	}

	extractedData := downloader.VideoData{
		Site:    "Facebook facebook.com",
		Title:   utils.FileName(title),
		Type:    "video",
		Formats: format,
	}
	extractedData.Download(url)
	return extractedData
}
