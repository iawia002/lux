package reddit

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
	"github.com/pkg/errors"
)

var (
	redditMP4API = "https://v.redd.it/"
	redditIMGAPI = "https://i.redd.it/"
	audioURLPart = "/DASH_audio.mp4"
	resURLParts  = []string{"/DASH_720.mp4", "/DASH_480.mp4", "/DASH_360.mp4", "/DASH_240.mp4", "/DASH_220.mp4"}
	res          = []string{"720p", "480p", "360p", "240p", "220p"}
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
	contentName := utils.MatchOneOf(html, `<title>(.+?)<\/title>`)[1]
	if utils.MatchOneOf(html, `meta property="og:video" content=.*HLSPlaylist`) != nil {
		fileType = "mp4"
	} else if utils.MatchOneOf(html, `<meta property="og:type" content="image"/>`) != nil {
		fileType = "img"
	} else if utils.MatchOneOf(html, `https:\/\/preview\.redd\.it\/.*gif`) != nil {
		fileType = "gif"
	}

	if fileType == "mp4" {
		mp4URL := utils.MatchOneOf(html, `https://v.redd.it/(.+?)/HLSPlaylist`)[1]
		if mp4URL == "" {
			return nil, errors.New("can't match mp4 content downloadable url")
		}

		streams := make(map[string]*extractors.Stream, len(resURLParts))
		for id := 0; id < len(resURLParts); id++ {
			index := strconv.Itoa(id)
			id, err := strconv.Atoi(index)
			if err != nil {
				return nil, err
			}
			resURL := fmt.Sprintf("%s%s%s", redditMP4API, mp4URL, resURLParts[id])
			audioURL := fmt.Sprintf("%s%s%s", redditMP4API, mp4URL, audioURLPart)
			vs, err := request.Size(resURL, referer)
			if err != nil {
				return nil, err
			}
			as, err := request.Size(audioURL, referer)
			if err != nil {
				return nil, err
			}
			parts := make([]*extractors.Part, 0, 2)
			parts = append(parts, &extractors.Part{
				URL:  resURL,
				Size: vs,
				Ext:  "mp4",
			})
			parts = append(parts, &extractors.Part{
				URL:  audioURL,
				Size: as,
				Ext:  "mp3",
			})
			var size int64
			for _, part := range parts {
				size += part.Size
			}
			streams[index] = &extractors.Stream{
				Parts:   parts,
				Size:    size,
				Quality: res[id],
				NeedMux: true,
			}
		}

		return []*extractors.Data{
			{
				Site:    "Reddit reddit.com",
				Title:   contentName,
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
		gifURL = strings.ReplaceAll(gifURL, "\"", "")

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
				Title:   contentName,
				Type:    extractors.DataTypeVideo,
				Streams: streams,
				URL:     url,
			},
		}, nil
	} else if fileType == "img" {
		var imgURL string
		var is int64
		if utils.MatchOneOf(html, `content":"https:\/\/i.redd.it\/(.+?)","type":"image"`) != nil {
			imgURL = redditIMGAPI + utils.MatchOneOf(html, `content":"https:\/\/i.redd.it\/(.+?)","type":"image"`)[1]
			is, err = request.Size(imgURL, referer)
			if err != nil {
				return nil, err
			}
		} else {
			// https://external-preview.redd.it/q3R30Ph7Bfh2kDgpX7jyvCru4NxRKvGcK3hYr2Ng3eM.jpg?auto=webp\u0026s=13131922131e99dff74b4f42179c11df3e091787
			// https://external-preview.redd.it/q3R30Ph7Bfh2kDgpX7jyvCru4NxRKvGcK3hYr2Ng3eM.jpg?auto=webp&s=13131922131e99dff74b4f42179c11df3e091787
			imgURL = utils.MatchOneOf(html, `content":"(.+?)","type":"image"`)[1]
			imgURL = strings.ReplaceAll(imgURL, "auto=webp\\u0026s", "auto=webp&s")
			is, err = request.Size(imgURL, referer)
			if err != nil {
				return nil, err
			}
		}

		imgParts := make([]*extractors.Part, 0)
		imgParts = append(imgParts, &extractors.Part{
			URL:  imgURL,
			Size: is,
			Ext:  "jpg",
		})

		return []*extractors.Data{
			{
				Site:  "Reddit reddit.com",
				Title: contentName,
				Type:  extractors.DataTypeVideo,
				Streams: map[string]*extractors.Stream{
					"image": {
						Parts: imgParts,
						Size:  is,
					},
				},
				URL: url,
			},
		}, nil
	}
	return nil, nil
}
