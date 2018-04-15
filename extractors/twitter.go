package extractors

import (
	"encoding/json"
	"html"
	"strings"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

const prefix = "https://twitter.com/i/videos/tweet/"

type twitter struct {
	VideoUrl string `json:"video_url"`
}

type twitterURLInfo struct {
	URL  string
	Size int64
}

// Twitter download function
func Twitter(uri string) downloader.VideoData {
	videoURI := getVideoURI(uri)
	e := download(videoURI, uri)
	e.Download(uri)
	return e
}

func getVideoURI(uri string) string {
	//extract tweet id from url
	tweetID := strings.Split(uri, "/")[5]
	webplayerURL := prefix + tweetID
	h := request.Get(webplayerURL, uri)
	//get dataconfig attribute
	jsonString := html.UnescapeString(utils.MatchOneOf(h, "data-config=\"({.+})")[1])
	var twitterData twitter
	//unmarshal
	json.Unmarshal([]byte(jsonString), &twitterData)
	return twitterData.VideoUrl
}

func download(directURI, uri string) downloader.VideoData {
	var size int64
	var urls []downloader.URLData
	switch {
	case strings.HasSuffix(directURI, "m3u8"):
		var m3u8URLs []twitterURLInfo
		m3u8URLs, size = twitterM3u8(directURI)

		var temp downloader.URLData
		for _, u := range m3u8URLs {
			temp = downloader.URLData{
				URL:  u.URL,
				Size: u.Size,
				Ext:  "ts",
			}
			urls = append(urls, temp)
		}

	case strings.HasSuffix(directURI, "mp4"):
		size = request.Size(directURI, uri)
		urlData := downloader.URLData{
			URL:  directURI,
			Size: size,
			Ext:  "mp4",
		}
		urls = []downloader.URLData{urlData}

	}
	format := map[string]downloader.FormatData{
		"default": {
			URLs: urls,
			Size: size,
		},
	}
	extractedData := downloader.VideoData{
		Site:    "Twitter twitter.com",
		Title:   "twitter_video",
		Type:    "video",
		Formats: format,
	}
	return extractedData
}

func twitterM3u8(uri string) ([]twitterURLInfo, int64) {
	var data []twitterURLInfo
	var temp twitterURLInfo
	var size, totalSize int64
	m3u8urls := utils.M3u8URLs(uri)
	var tsurls []string
	for _, u := range m3u8urls {
		tsurls = append(tsurls, utils.M3u8URLs(u)...)
	}
	for _, u := range tsurls {
		size = request.Size(u, uri)
		totalSize += size
		temp = twitterURLInfo{
			URL:  u,
			Size: size,
		}
		data = append(data, temp)
	}
	return data, totalSize
}
