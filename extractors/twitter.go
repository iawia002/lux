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
	VideoURL string `json:"video_url"`
}

type twitterURLInfo struct {
	URLs []string
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
	return twitterData.VideoURL
}

//download func
func download(directURI, uri string) downloader.VideoData {
	var size int64
	var format = make(map[string]downloader.FormatData)
	switch {
	//if video file is m3u8 and ts
	case strings.HasSuffix(directURI, "m3u8"):
		vInfo := UnifyVideoFiles(directURI)
		vInfoNum := len(vInfo)
		counter := 1
		for _, u := range vInfo {
			var urls []downloader.URLData
			for _, i := range u.URLs {
				temp := downloader.URLData{
					URL:  i,
					Size: u.Size,
					Ext:  "ts",
				}
				urls = append(urls, temp)
			}
			quality := strings.Split(u.URLs[0], "/")[9]
			if counter == vInfoNum {
				format["default"] = downloader.FormatData{Quality: quality, URLs: urls, Size: u.Size}
			} else {
				format[quality] = downloader.FormatData{Quality: quality, URLs: urls, Size: u.Size}
			}
			counter++
		}

		//if video file is mp4
	case strings.HasSuffix(directURI, "mp4"):
		size = request.Size(directURI, uri)
		urlData := downloader.URLData{
			URL:  directURI,
			Size: size,
			Ext:  "mp4",
		}
		format["default"] = downloader.FormatData{URLs: []downloader.URLData{urlData}, Size: size}
	}

	extractedData := downloader.VideoData{
		Site:    "Twitter twitter.com",
		Title:   "twitter_video",
		Type:    "video",
		Formats: format,
	}
	return extractedData
}

//UnifyVideoFiles unify files infomation
func UnifyVideoFiles(uri string) []twitterURLInfo {
	var data []twitterURLInfo
	m3u8urls := utils.M3u8URLs(uri)
	for _, u := range m3u8urls {
		var totalSize int64
		ts := utils.M3u8URLs(u)
		for _, i := range ts {
			size := request.Size(i, uri)
			totalSize += size
		}
		temp := twitterURLInfo{
			URLs: ts,
			Size: totalSize,
		}
		data = append(data, temp)
	}
	return data
}
