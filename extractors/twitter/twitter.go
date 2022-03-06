package twitter

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("twitter", New())
}

type twitter struct {
	Track struct {
		URL string `json:"playbackUrl"`
	} `json:"track"`
	TweetID  string
	Username string
}

type extractor struct{}

// New returns a twitter extractor.
func New() extractors.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	usernames := utils.MatchOneOf(html, `property="og:title"\s+content="(.+)"`)
	if usernames == nil || len(usernames) < 2 {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}
	username := usernames[1]

	tweetIDs := utils.MatchOneOf(url, `(status|statuses)/(\d+)`)
	if tweetIDs == nil || len(tweetIDs) < 3 {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}
	tweetID := tweetIDs[2]

	api := fmt.Sprintf(
		"https://api.twitter.com/1.1/videos/tweet/config/%s.json", tweetID,
	)
	headers := map[string]string{
		"Authorization": "Bearer AAAAAAAAAAAAAAAAAAAAAIK1zgAAAAAA2tUWuhGZ2JceoId5GwYWU5GspY4%3DUq7gzFoCZs1QfwGoVdvSac3IniczZEYXIcDyumCauIXpcAPorE",
	}
	jsonString, err := request.Get(api, url, headers)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var twitterData twitter
	if err := json.Unmarshal([]byte(jsonString), &twitterData); err != nil {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}
	twitterData.TweetID = tweetID
	twitterData.Username = username
	extractedData, err := download(twitterData, url)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return extractedData, nil
}

func download(data twitter, uri string) ([]*extractors.Data, error) {
	var (
		err  error
		size int64
	)
	streams := make(map[string]*extractors.Stream)
	switch {
	// if video file is m3u8 and ts
	case strings.Contains(data.Track.URL, ".m3u8"):
		m3u8urls, err := utils.M3u8URLs(data.Track.URL)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		for index, m3u8 := range m3u8urls {
			var totalSize int64
			ts, err := utils.M3u8URLs(m3u8)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			urls := make([]*extractors.Part, 0, len(ts))
			for _, i := range ts {
				size, err := request.Size(i, uri)
				if err != nil {
					return nil, errors.WithStack(err)
				}
				temp := &extractors.Part{
					URL:  i,
					Size: size,
					Ext:  "ts",
				}
				totalSize += size
				urls = append(urls, temp)
			}
			qualityString := utils.MatchOneOf(m3u8, `/(\d+x\d+)/`)[1]
			quality := strconv.Itoa(index + 1)
			streams[quality] = &extractors.Stream{
				Parts:   urls,
				Size:    totalSize,
				Quality: qualityString,
			}
		}

	// if video file is mp4
	case strings.Contains(data.Track.URL, ".mp4"):
		size, err = request.Size(data.Track.URL, uri)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		urlData := &extractors.Part{
			URL:  data.Track.URL,
			Size: size,
			Ext:  "mp4",
		}
		streams["default"] = &extractors.Stream{
			Parts: []*extractors.Part{urlData},
			Size:  size,
		}
	}

	return []*extractors.Data{
		{
			Site:    "Twitter twitter.com",
			Title:   fmt.Sprintf("%s %s", data.Username, data.TweetID),
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     uri,
		},
	}, nil
}
