package extractors

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type twitter struct {
	Track struct {
		URL string `json:"playbackUrl"`
	} `json:"track"`
	TweetID  string
	Username string
}

// Twitter download function
func Twitter(uri string) downloader.VideoData {
	html := request.Get(uri, uri, nil)
	username := utils.MatchOneOf(html, `property="og:title"\s+content="(.+)"`)[1]
	tweetID := utils.MatchOneOf(uri, `(status|statuses)/(\d+)`)[2]
	api := fmt.Sprintf(
		"https://api.twitter.com/1.1/videos/tweet/config/%s.json", tweetID,
	)
	headers := map[string]string{
		"Authorization": "Bearer AAAAAAAAAAAAAAAAAAAAAIK1zgAAAAAA2tUWuhGZ2JceoId5GwYWU5GspY4%3DUq7gzFoCZs1QfwGoVdvSac3IniczZEYXIcDyumCauIXpcAPorE",
	}
	jsonString := request.Get(api, uri, headers)
	var twitterData twitter
	json.Unmarshal([]byte(jsonString), &twitterData)
	twitterData.TweetID = tweetID
	twitterData.Username = username
	extractedData := download(twitterData, uri)
	extractedData.Download(uri)
	return extractedData
}

func download(data twitter, uri string) downloader.VideoData {
	var size int64
	var format = make(map[string]downloader.FormatData)
	switch {
	// if video file is m3u8 and ts
	case strings.HasSuffix(data.Track.URL, "m3u8"):
		m3u8urls := utils.M3u8URLs(data.Track.URL)
		for index, m3u8 := range m3u8urls {
			var totalSize int64
			var urls []downloader.URLData
			ts := utils.M3u8URLs(m3u8)
			for _, i := range ts {
				size := request.Size(i, uri)
				temp := downloader.URLData{
					URL:  i,
					Size: size,
					Ext:  "ts",
				}
				totalSize += size
				urls = append(urls, temp)
			}
			qualityString := utils.MatchOneOf(m3u8, `/(\d+x\d+)/`)[1]
			quality := strconv.Itoa(index + 1)
			if index+1 == len(m3u8urls) {
				quality = "default"
			}
			format[quality] = downloader.FormatData{
				Quality: qualityString,
				URLs:    urls,
				Size:    totalSize,
			}
		}

	// if video file is mp4
	case strings.HasSuffix(data.Track.URL, "mp4"):
		size = request.Size(data.Track.URL, uri)
		urlData := downloader.URLData{
			URL:  data.Track.URL,
			Size: size,
			Ext:  "mp4",
		}
		format["default"] = downloader.FormatData{
			URLs: []downloader.URLData{urlData},
			Size: size,
		}
	}

	extractedData := downloader.VideoData{
		Site:    "Twitter twitter.com",
		Title:   fmt.Sprintf("%s %s", data.Username, data.TweetID),
		Type:    "video",
		Formats: format,
	}
	return extractedData
}
