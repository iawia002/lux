package twitter

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/iawia002/annie/extractors/types"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type twitter struct {
	Track struct {
		URL string `json:"playbackUrl"`
	} `json:"track"`
	TweetID  string
	Username string
}

type extractor struct{}

// New returns a youtube extractor.
func New() types.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option types.Options) ([]*types.Data, error) {
	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, err
	}

	usernames := utils.MatchOneOf(html, `property="og:title"\s+content="(.+)"`)
	if usernames == nil || len(usernames) < 2 {
		return nil, types.ErrURLParseFailed
	}
	username := usernames[1]

	tweetIDs := utils.MatchOneOf(url, `(status|statuses)/(\d+)`)
	if tweetIDs == nil || len(tweetIDs) < 3 {
		return nil, types.ErrURLParseFailed
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
		return nil, err
	}

	var twitterData twitter
	if err := json.Unmarshal([]byte(jsonString), &twitterData); err != nil {
		return nil, types.ErrURLParseFailed
	}
	twitterData.TweetID = tweetID
	twitterData.Username = username
	extractedData, err := download(twitterData, url)
	if err != nil {
		return nil, err
	}
	return extractedData, nil
}

func download(data twitter, uri string) ([]*types.Data, error) {
	var (
		err  error
		size int64
	)
	streams := make(map[string]*types.Stream)
	switch {
	// if video file is m3u8 and ts
	case strings.Contains(data.Track.URL, ".m3u8"):
		m3u8urls, err := utils.M3u8URLs(data.Track.URL)
		if err != nil {
			return nil, err
		}
		for index, m3u8 := range m3u8urls {
			var totalSize int64
			ts, err := utils.M3u8URLs(m3u8)
			if err != nil {
				return nil, err
			}
			urls := make([]*types.Part, 0, len(ts))
			for _, i := range ts {
				size, err := request.Size(i, uri)
				if err != nil {
					return nil, err
				}
				temp := &types.Part{
					URL:  i,
					Size: size,
					Ext:  "ts",
				}
				totalSize += size
				urls = append(urls, temp)
			}
			qualityString := utils.MatchOneOf(m3u8, `/(\d+x\d+)/`)[1]
			quality := strconv.Itoa(index + 1)
			streams[quality] = &types.Stream{
				Parts:   urls,
				Size:    totalSize,
				Quality: qualityString,
			}
		}

	// if video file is mp4
	case strings.Contains(data.Track.URL, ".mp4"):
		size, err = request.Size(data.Track.URL, uri)
		if err != nil {
			return nil, err
		}
		urlData := &types.Part{
			URL:  data.Track.URL,
			Size: size,
			Ext:  "mp4",
		}
		streams["default"] = &types.Stream{
			Parts: []*types.Part{urlData},
			Size:  size,
		}
	}

	return []*types.Data{
		{
			Site:    "Twitter twitter.com",
			Title:   fmt.Sprintf("%s %s", data.Username, data.TweetID),
			Type:    types.DataTypeVideo,
			Streams: streams,
			URL:     uri,
		},
	}, nil
}
