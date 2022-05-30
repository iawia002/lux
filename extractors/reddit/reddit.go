package reddit

import (
	"fmt"
	"log"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
	"github.com/pkg/errors"
)

var (
	PostId       string
	redditMP4API = "https://v.redd.it/"
	audioURLPart = "/DASH_audio.mp4"
	res720       = "/DASH_720.mp4"
	res480       = "/DASH_480.mp4"
	res360       = "/DASH_360.mp4"
	res280       = "/DASH_280.mp4"
)

type extractor struct{}

const referer = "https://www.reddit.com"

func init() {
	extractors.Register("reddit", New())
}

func New() extractors.Extractor {
	return &extractor{}
}

func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	var err error
	html, err := request.Get(url, referer, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// set thread number to 1 manually to avoid http 412 error
	option.ThreadNumber = 1

	return extractContent(url, html, option), err
}

func extractContent(url, html string, extractOption extractors.Options) []*extractors.Data {

	var fileType = ""
	videoName := utils.MatchOneOf(html, `<title>.*<\/title>`)[0]
	if utils.MatchOneOf(html, `meta property="og:video" content=.*HLSPlaylist`) != nil {
		fileType = "mp4"
	} else if utils.MatchOneOf(html, `https:\/\/preview\.redd\.it\/.*gif`) != nil {
		fileType = "gif"
	}

	if fileType == "mp4" {
		url := utils.MatchOneOf(html, `https://v.redd.it/.*/HLSPlaylist`)[0]
		if url == "" {
			panic("can't match anything")
		}

		for i := len(url) - 1; i >= 0; i-- {
			if url[i] == '/' {
				url = url[:i]
				break
			}
		}
		for i := len(url) - 1; i >= 0; i-- {
			if url[i] == '/' {
				url = url[i+1:]
				break
			}
		}

		for i := len(videoName) - 1; i >= 0; i-- {
			if videoName[i] == '<' {
				videoName = videoName[:i]
				break
			}
		}

		for i := len(videoName) - 1; i >= 0; i-- {
			if videoName[i] == '>' {
				videoName = videoName[i+1:]
				break
			}
		}

		videoURL := fmt.Sprintf("%s%s%s", redditMP4API, url, res720)
		// audioURL := fmt.Sprintf("%s%s%s", redditMP4API, url, audioURLPart)
		videoSize, err := request.Size(videoURL, "reddit.com")
		if err != nil {
			log.Fatal("can't get video size")
		}
		// audioSize, err := request.Size(audioURL, "reddit.com")
		// if err != nil {
		// 	log.Fatal("can't get video size")
		// }

		contentData := make([]*extractors.Part, 0)
		contentData = append(contentData, &extractors.Part{
			URL:  videoURL,
			Size: videoSize,
			Ext:  "mp4",
		})
		// contentData = append(contentData, &extractors.Part{
		// 	URL:  audioURL,
		// 	Size: audioSize,
		// 	Ext:  "mp4",
		// })

		streams := map[string]*extractors.Stream{
			"default": {
				Parts: contentData,
				Size:  videoSize,
				// Size:  videoSize + audioSize,
			},
		}

		return []*extractors.Data{
			{
				Site:    "reddit",
				Title:   videoName,
				Type:    extractors.DataTypeVideo,
				Streams: streams,
				URL:     "www.reddit.com",
			},
		}
	} else if fileType == "gif" {
		url, urlU, urlL := "", "", ""
		urls := utils.MatchOneOf(html, `https:\/\/preview\.redd\.it\/.*?\.gif\?format=mp4.*?"`)
		if urls != nil {
			url = urls[0]
		}
		if url == "" {
			panic("can't match anything")
		}

		for i := len(url) - 1; i >= 0; i-- {
			if url[i] == '&' {
				urlU = url[:i+1]
				break
			}
		}

		for i := len(url) - 1; i >= 0; i-- {
			if url[i] == ';' {
				urlL = url[i+1 : len(url)-1]
				break
			}
		}

		url = urlU + urlL

		for i := len(videoName) - 1; i >= 0; i-- {
			if videoName[i] == '<' {
				videoName = videoName[:i]
				break
			}
		}

		for i := len(videoName) - 1; i >= 0; i-- {
			if videoName[i] == '>' {
				videoName = videoName[i+1:]
				break
			}
		}

		url = fmt.Sprintf("%s", url)

		videoSize, err := request.Size(url, "reddit.com")
		if err != nil {
			log.Fatal("can't get video size")
		}
		contentData := make([]*extractors.Part, 0)
		contentData = append(contentData, &extractors.Part{
			URL:  url,
			Size: videoSize,
			Ext:  "mp4",
		})

		streams := map[string]*extractors.Stream{
			"video": {
				Parts: contentData,
				Size:  videoSize,
			},
		}

		return []*extractors.Data{
			{
				Site:    "reddit",
				Title:   videoName,
				Type:    extractors.DataTypeVideo,
				Streams: streams,
				URL:     url,
			},
		}
	}
	return nil
}
