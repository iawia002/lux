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
	Status   struct {
		Text string `json:"text"`
	} `json:"status"`
}

type twitterURLInfo struct {
	URLs []string
	Size int64
}

// Twitter download function
func Twitter(uri string) downloader.VideoData {
	twitterData := getVideoURI(uri)
	extractedData := download(twitterData.VideoURL, uri, twitterData.Status.Text)
	extractedData.Download(uri)
	return extractedData
}

func getVideoURI(uri string) twitter {
	//extract tweet id from url
	tweetID := utils.MatchOneOf(uri, `/(\d+)`)[1]
	webplayerURL := prefix + tweetID
	h := request.Get(webplayerURL, uri)
	//get dataconfig attribute
	jsonString := html.UnescapeString(utils.MatchOneOf(h, `data-config="({.+})`)[1])
	var twitterData twitter
	//unmarshal
	json.Unmarshal([]byte(jsonString), &twitterData)
	waste := utils.MatchOneOf(twitterData.Status.Text, `https://t.co/\S+$`)[0]
	// twitterData.Status.Text has video url at end of the text. delete it!
	twitterData.Status.Text = twitterData.Status.Text[:len(twitterData.Status.Text)-len(waste)]
	//Sometimes, twitterData.Status.Text has newline character(\n). It prevent FFMPEG from merging files,so replace it with "".
	twitterData.Status.Text = strings.Replace(twitterData.Status.Text, "\n", "", -1)
	//if tweet has video only
	if twitterData.Status.Text == "" {
		twitterData.Status.Text = "twitter_video"
	} else {
		twitterData.Status.Text = twitterData.Status.Text[:len(twitterData.Status.Text)-1]
	}
	return twitterData
}

//download func
func download(directURI, uri, title string) downloader.VideoData {
	var size int64
	var format = make(map[string]downloader.FormatData)
	switch {
	//if video file is m3u8 and ts
	case strings.HasSuffix(directURI, "m3u8"):
		var vInfo []twitterURLInfo
		m3u8urls := utils.M3u8URLs(directURI)
		for _, u := range m3u8urls {
			var totalSize int64
			ts := utils.M3u8URLs(u)
			for _, i := range ts {
				size := request.Size(i, directURI)
				totalSize += size
			}
			temp := twitterURLInfo{
				URLs: ts,
				Size: totalSize,
			}
			vInfo = append(vInfo, temp)
		}
		vInfoNum := len(vInfo)
		for index, u := range vInfo {
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
			if index+1 == vInfoNum {
				format["default"] = downloader.FormatData{Quality: quality, URLs: urls, Size: u.Size}
			} else {
				format[quality] = downloader.FormatData{Quality: quality, URLs: urls, Size: u.Size}
			}
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
		Title:   utils.FileName(title),
		Type:    "video",
		Formats: format,
	}
	return extractedData
}
