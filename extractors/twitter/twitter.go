package twitter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	netURL "net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

const (
	twitterVideoEndpoint = "https://api.twitter.com/1.1/videos/tweet/config/%s.json"
	xquikTweetEndpoint   = "https://xquik.com/api/v1/x/tweets/%s"
	xquikAPIKeyEnv       = "XQUIK_API_KEY"
	maxXquikResponseSize = 4 << 20
)

func init() {
	e := New()
	extractors.Register("twitter", e)
	extractors.Register("x", e)
}

type twitter struct {
	Track struct {
		URL string `json:"playbackUrl"`
	} `json:"track"`
	TweetID  string
	Username string
}

type extractor struct {
	twitterEndpoint string
	xquikEndpoint   string
	client          *http.Client
}

// New returns a twitter extractor.
func New() extractors.Extractor {
	return &extractor{
		twitterEndpoint: twitterVideoEndpoint,
		xquikEndpoint:   xquikTweetEndpoint,
		client: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	tweetIDs := utils.MatchOneOf(url, `(status|statuses)/(\d+)`)
	if len(tweetIDs) < 3 {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}
	tweetID := tweetIDs[2]

	if apiKey := strings.TrimSpace(os.Getenv(xquikAPIKeyEnv)); apiKey != "" {
		return e.extractWithXquik(url, tweetID, apiKey)
	}

	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	username := twitterTitle(html, url)

	api := fmt.Sprintf(e.twitterEndpoint, tweetID)
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

func twitterTitle(html, uri string) string {
	if titles := utils.MatchOneOf(html, `property="og:title"\s+content="(.+)"`); len(titles) >= 2 {
		return titles[1]
	}

	parsed, err := netURL.Parse(uri)
	if err != nil {
		return "Twitter"
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) >= 3 && (parts[1] == "status" || parts[1] == "statuses") && parts[0] != "i" {
		return "@" + parts[0]
	}
	return "Twitter"
}

type xquikResponse struct {
	Tweet struct {
		Author *xquikAuthor `json:"author"`
		Media  []xquikMedia `json:"media"`
	} `json:"tweet"`
	Author *xquikAuthor `json:"author"`
}

type xquikAuthor struct {
	Name     string `json:"name"`
	Username string `json:"username"`
}

type xquikMedia struct {
	Type          string              `json:"type"`
	VideoVariants []xquikVideoVariant `json:"videoVariants"`
}

type xquikVideoVariant struct {
	Bitrate     int    `json:"bitrate"`
	ContentType string `json:"contentType"`
	URL         string `json:"url"`
}

func (e *extractor) extractWithXquik(uri, tweetID, apiKey string) ([]*extractors.Data, error) {
	api := fmt.Sprintf(e.xquikEndpoint, tweetID)
	jsonBytes, err := e.getXquikTweet(api, apiKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var response xquikResponse
	if err := json.Unmarshal(jsonBytes, &response); err != nil {
		return nil, errors.WithStack(extractors.ErrBodyParseFailed)
	}

	author := response.Author
	if response.Tweet.Author != nil {
		author = response.Tweet.Author
	}
	title := xquikTitle(author, tweetID)
	videos := make([]*extractors.Data, 0, len(response.Tweet.Media))
	for _, media := range response.Tweet.Media {
		if media.Type != "video" && media.Type != "animated_gif" {
			continue
		}
		streams, err := xquikStreams(media.VideoVariants, uri)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if len(streams) == 0 {
			continue
		}
		videos = append(videos, &extractors.Data{
			Site:    "X (Twitter) x.com",
			Title:   title,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     uri,
		})
	}

	if len(videos) == 0 {
		return nil, errors.New("tweet has no downloadable MP4 video")
	}
	if len(videos) > 1 {
		for index := range videos {
			videos[index].Title = fmt.Sprintf("%s [%d]", title, index+1)
		}
	}
	return videos, nil
}

func (e *extractor) getXquikTweet(api, apiKey string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, api, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	req.Header.Set("x-api-key", apiKey)

	res, err := e.client.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer res.Body.Close() // nolint
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return nil, errors.Errorf("%s request error: HTTP %d", api, res.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(res.Body, maxXquikResponseSize+1))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if len(body) > maxXquikResponseSize {
		return nil, errors.New("Xquik response exceeds 4 MiB")
	}
	return body, nil
}

func xquikTitle(author *xquikAuthor, tweetID string) string {
	if author == nil {
		return fmt.Sprintf("X video %s", tweetID)
	}
	name := strings.TrimSpace(author.Name)
	username := strings.TrimPrefix(strings.TrimSpace(author.Username), "@")
	switch {
	case name != "" && username != "":
		return fmt.Sprintf("%s (@%s) %s", name, username, tweetID)
	case name != "":
		return fmt.Sprintf("%s %s", name, tweetID)
	case username != "":
		return fmt.Sprintf("@%s %s", username, tweetID)
	default:
		return fmt.Sprintf("X video %s", tweetID)
	}
}

func xquikStreams(variants []xquikVideoVariant, referer string) (map[string]*extractors.Stream, error) {
	streams := make(map[string]*extractors.Stream)
	seenURLs := make(map[string]struct{})
	for _, variant := range variants {
		if !strings.HasPrefix(strings.ToLower(variant.ContentType), "video/mp4") || variant.URL == "" {
			continue
		}
		if _, exists := seenURLs[variant.URL]; exists {
			continue
		}
		seenURLs[variant.URL] = struct{}{}

		size, err := request.Size(variant.URL, referer)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		id := "default"
		quality := "MP4"
		if variant.Bitrate > 0 {
			id = strconv.Itoa(variant.Bitrate)
			quality = fmt.Sprintf("%d kbps MP4", variant.Bitrate/1000)
		}
		if _, exists := streams[id]; exists {
			id = fmt.Sprintf("%s-%d", id, len(streams)+1)
		}
		streams[id] = &extractors.Stream{
			Parts: []*extractors.Part{{
				URL:  variant.URL,
				Size: size,
				Ext:  "mp4",
			}},
			Size:    size,
			Quality: quality,
		}
	}
	return streams, nil
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
