package reddit

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("reddit", New())
}

const (
	referer  = "https://www.reddit.com"
	siteName = "Reddit reddit.com"

	redditMP4API = "https://v.redd.it/"
	redditIMGAPI = "https://i.redd.it/"
	audioURLPart = "/DASH_audio.mp4"
)

var resMap = map[string]string{
	"720p": "/DASH_720.mp4",
	"480p": "/DASH_480.mp4",
	"360p": "/DASH_360.mp4",
	"240p": "/DASH_240.mp4",
	"220p": "/DASH_220.mp4",
}

type extractor struct{}

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

	title := utils.MatchOneOf(html, `<title>(.+?)<\/title>`)[1]

	if utils.MatchOneOf(html, `meta property="og:video" content=.*HLSPlaylist`) != nil {
		mp4URL := utils.MatchOneOf(html, `https://v.redd.it/(.+?)/HLSPlaylist`)[1]
		if mp4URL == "" {
			return nil, errors.New("can't match mp4 content downloadable url")
		}

		audioURL := fmt.Sprintf("%s%s%s", redditMP4API, mp4URL, audioURLPart)
		size, err := request.Size(audioURL, referer)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		audioPart := &extractors.Part{
			URL:  audioURL,
			Size: size,
			Ext:  "mp3",
		}

		streams := make(map[string]*extractors.Stream, len(resMap))
		for res, urlParts := range resMap {
			resURL := fmt.Sprintf("%s%s%s", redditMP4API, mp4URL, urlParts)
			size, err := request.Size(resURL, referer)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			streams[res] = &extractors.Stream{
				Parts: []*extractors.Part{
					{
						URL:  resURL,
						Size: size,
						Ext:  "mp4",
					},
					audioPart,
				},
				Size:    size + audioPart.Size,
				Quality: res,
				NeedMux: true,
			}
		}

		return []*extractors.Data{
			{
				Site:    siteName,
				Title:   title,
				Type:    extractors.DataTypeVideo,
				Streams: streams,
				URL:     url,
			},
		}, nil
	} else if utils.MatchOneOf(html, `<meta property="og:type" content="image"/>`) != nil {
		var imgURL string
		var size int64
		if utils.MatchOneOf(html, `content":"https:\/\/i.redd.it\/(.+?)","type":"image"`) != nil {
			imgURL = redditIMGAPI + utils.MatchOneOf(html, `content":"https:\/\/i.redd.it\/(.+?)","type":"image"`)[1]
			size, err = request.Size(imgURL, referer)
			if err != nil {
				return nil, errors.WithStack(err)
			}
		} else {
			imgURL = utils.MatchOneOf(html, `content":"(.+?)","type":"image"`)[1]
			imgURL = strings.ReplaceAll(imgURL, "auto=webp\\u0026s", "auto=webp&s")
			size, err = request.Size(imgURL, referer)
			if err != nil {
				return nil, errors.WithStack(err)
			}
		}

		return []*extractors.Data{
			{
				Site:  siteName,
				Title: title,
				Type:  extractors.DataTypeImage,
				Streams: map[string]*extractors.Stream{
					"default": {
						Parts: []*extractors.Part{
							{
								URL:  imgURL,
								Size: size,
								Ext:  "jpg",
							},
						},
						Size: size,
					},
				},
				URL: url,
			},
		}, nil
	} else if utils.MatchOneOf(html, `https:\/\/preview\.redd\.it\/.*gif`) != nil {
		gifURL := utils.MatchOneOf(html, `https:\/\/preview\.redd\.it\/.*?\.gif\?format=mp4.*?"`)[0]
		if gifURL == "" {
			return nil, errors.New("can't match gif content downloadable url")
		}

		gifURL = strings.ReplaceAll(gifURL, "&amp;", "&")
		gifURL = strings.ReplaceAll(gifURL, "\"", "")

		size, err := request.Size(gifURL, "reddit.com")
		if err != nil {
			return nil, errors.New("can't get video size")
		}

		streams := map[string]*extractors.Stream{
			"default": {
				Parts: []*extractors.Part{
					{
						URL:  gifURL,
						Size: size,
						Ext:  "mp4",
					},
				},
				Size: size,
			},
		}
		return []*extractors.Data{
			{
				Site:    siteName,
				Title:   title,
				Type:    extractors.DataTypeVideo,
				Streams: streams,
				URL:     url,
			},
		}, nil
	}

	return nil, fmt.Errorf("unable to handle url: %s", url)
}
