package udn

import (
	"errors"
	"strings"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/extractors"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

const (
	startFlag = `',
            mp4: '//`
	endFlag = `'
        },
        subtitles`
)

func getCDNUrl(html string) string {
	if cdnURLs := utils.MatchOneOf(html, startFlag+"(.+?)"+endFlag); cdnURLs != nil && len(cdnURLs) > 1 && cdnURLs[1] != "" {
		return cdnURLs[1]
	}
	return ""
}

func prepareEmbedURL(url string) string {
	if !strings.Contains(url, "https://video.udn.com/embed/") {
		newIDs := strings.Split(url, "/")
		if len(newIDs) < 1 {
			return ""
		}
		return "https://video.udn.com/embed/news/" + newIDs[len(newIDs)-1]
	}
	return url
}

// Extract is the main function for extracting data
func Extract(url string) ([]downloader.Data, error) {
	url = prepareEmbedURL(url)
	if len(url) == 0 {
		return nil, extractors.ErrURLParseFailed
	}

	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, err
	}
	var title string
	desc := utils.MatchOneOf(html, `title: '(.+?)',
        link:`)
	if desc != nil && len(desc) > 1 {
		title = desc[1]
	} else {
		title = "udn"
	}
	cdnURL := getCDNUrl(html)
	if cdnURL == "" {
		return nil, errors.New("empty list")
	}
	srcURL, err := request.Get("http://"+cdnURL, url, nil)
	if err != nil {
		return nil, err
	}
	size, err := request.Size(srcURL, url)
	if err != nil {
		return nil, err
	}
	urlData := downloader.URL{
		URL:  srcURL,
		Size: size,
		Ext:  "mp4",
	}
	quality := "normal"
	streams := map[string]downloader.Stream{
		quality: {
			URLs:    []downloader.URL{urlData},
			Size:    size,
			Quality: quality,
		},
	}
	return []downloader.Data{
		{
			Site:    "udn udn.com",
			Title:   title,
			Type:    "video",
			Streams: streams,
			URL:     url,
		},
	}, nil
}
