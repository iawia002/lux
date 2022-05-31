package reddit

import (
	"fmt"
	"strings"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
	"github.com/pkg/errors"
)

var (
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
	html, err := request.Get(url, referer, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// set thread number to 1 manually to avoid http 412 error
	option.ThreadNumber = 1

	var fileType = ""
	videoName := utils.MatchOneOf(html, `<title>(.+?)<\/title>`)[1]
	if utils.MatchOneOf(html, `meta property="og:video" content=.*HLSPlaylist`) != nil {
		fileType = "mp4"
	} else if utils.MatchOneOf(html, `https:\/\/preview\.redd\.it\/.*gif`) != nil {
		fileType = "gif"
	}

	if fileType == "mp4" {
		mp4URL := utils.MatchOneOf(html, `https://v.redd.it/(.+?)/HLSPlaylist`)[1]
		if mp4URL == "" {
			return nil, errors.New("can't match mp4 content downloadable url")
		}

		videoURL := fmt.Sprintf("%s%s%s", redditMP4API, mp4URL, res720)
		audioURL := fmt.Sprintf("%s%s%s", redditMP4API, mp4URL, audioURLPart)
		videoSize, err := request.Size(videoURL, "reddit.com")
		if err != nil {
			return nil, err
		}
		audioSize, err := request.Size(audioURL, "reddit.com")
		if err != nil {
			return nil, err
		}

		contentData := make([]*extractors.Part, 0)
		contentData = append(contentData, &extractors.Part{
			URL:  videoURL,
			Size: videoSize,
			Ext:  "mp4",
		})

		contentData = append(contentData,
			&extractors.Part{
				URL:  audioURL,
				Size: audioSize,
				Ext:  "mp3",
			})

		streams := map[string]*extractors.Stream{
			"default": {
				Parts:   contentData,
				Size:    videoSize + audioSize,
				NeedMux: true,
			},
		}

		return []*extractors.Data{
			{
				Site:    "Reddit reddit.com",
				Title:   videoName,
				Type:    extractors.DataTypeVideo,
				Streams: streams,
				URL:     url,
			},
		}, nil
	} else if fileType == "gif" {
		gifURL := utils.MatchOneOf(html, `https:\/\/preview\.redd\.it\/.*?\.gif\?format=mp4.*?"`)[0]
		if gifURL == "" {
			return nil, errors.New("can't match gif content downloadable url")
		}

		gifURL = strings.ReplaceAll(gifURL, "&amp;", "&")

		videoSize, err := request.Size(gifURL, "reddit.com")
		if err != nil {
			return nil, errors.New("can't get video size")
		}
		contentData := make([]*extractors.Part, 0)
		contentData = append(contentData, &extractors.Part{
			URL:  gifURL,
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
				Site:    "Reddit reddit.com",
				Title:   videoName,
				Type:    extractors.DataTypeVideo,
				Streams: streams,
				URL:     url,
			},
		}, nil
	}
	return nil, nil
}
