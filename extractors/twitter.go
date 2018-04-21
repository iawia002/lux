package extractors

import (
	"encoding/json"
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

const prefix = "https://twitter.com/i/videos/tweet/"

type twitterUser struct {
	Name string `json:"name"`
}

type twitter struct {
	VideoURL string      `json:"video_url"`
	TweetID  string      `json:"tweet_id"`
	User     twitterUser `json:"user"`
}

// Twitter download function
func Twitter(uri string) downloader.VideoData {
	twitterData := getVideoURI(uri)
	extractedData := download(twitterData, uri)
	extractedData.Download(uri)
	return extractedData
}

func getVideoURI(uri string) twitter {
	// extract tweet id from url
	tweetID := utils.MatchOneOf(uri, `(status|statuses)/(\d+)`)[2]
	webPlayerURL := prefix + tweetID
	h := request.Get(webPlayerURL, uri)
	// get dataconfig attribute
	jsonString := html.UnescapeString(utils.MatchOneOf(h, `data-config="({.+})`)[1])
	var twitterData twitter
	json.Unmarshal([]byte(jsonString), &twitterData)
	return twitterData
}

func download(data twitter, uri string) downloader.VideoData {
	var size int64
	var format = make(map[string]downloader.FormatData)
	switch {
	// if video file is m3u8 and ts
	case strings.HasSuffix(data.VideoURL, "m3u8"):
		m3u8urls := utils.M3u8URLs(data.VideoURL)
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
	case strings.HasSuffix(data.VideoURL, "mp4"):
		size = request.Size(data.VideoURL, uri)
		urlData := downloader.URLData{
			URL:  data.VideoURL,
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
		Title:   fmt.Sprintf("%s %s", data.User.Name, data.TweetID),
		Type:    "video",
		Formats: format,
	}
	return extractedData
}
