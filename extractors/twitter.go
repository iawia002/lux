package extractors

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/url"
	"strings"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

const prefix = "https://twitter.com/i/videos/tweet/"

type twitter struct {
	VideoUrl string `json:"video_url"`
}

// Twitter download function
func Twitter(uri string) {
	videoURI := getVideoURI(uri)
	vu, err := url.Parse(videoURI)
	if err != nil {
		log.Println(err)
	}
	ep := vu.EscapedPath()
	switch {
	case strings.HasSuffix(ep, "m3u8"):
		fmt.Println(utils.M3u8URLs(videoURI))
	case strings.HasSuffix(ep, "mp4"):
		download(videoURI, uri)
	}
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

func download(directURI, uri string) {
	size := request.Size(directURI, uri)
	urlData := downloader.URLData{
		URL:  directURI,
		Size: size,
		Ext:  "mp4",
	}
	format := map[string]downloader.FormatData{
		"default": {
			URLs: []downloader.URLData{urlData},
			Size: size,
		},
	}
	extractedData := downloader.VideoData{
		Site:    "Twitter twitter.com",
		Title:   "twitter_video",
		Type:    "video",
		Formats: format,
	}
	extractedData.Download(uri)
}
