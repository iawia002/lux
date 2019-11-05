package geekbang

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/extractors"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type geekData struct {
	Code  int             `json:"code"`
	Error json.RawMessage `json:"error"`
	Data  struct {
		Title         string `json:"article_sharetitle"`
		VideoMediaMap map[string]struct {
			URL  string `json:"url"`
			Size int64  `json:"size"`
		} `json:"video_media_map"`
	} `json:"data"`
}

type geekURLInfo struct {
	URL  string
	Size int64
}

func geekM3u8(url string) ([]geekURLInfo, error) {
	var (
		data []geekURLInfo
		temp geekURLInfo
		size int64
		err  error
	)
	urls, err := utils.M3u8URLs(url)
	if err != nil {
		return nil, err
	}
	for _, u := range urls {
		temp = geekURLInfo{
			URL:  u,
			Size: size,
		}
		data = append(data, temp)
	}
	return data, nil
}

// Extract is the main function for extracting data
func Extract(url string) ([]downloader.Data, error) {
	var err error
	matches := utils.MatchOneOf(url, `https?://time.geekbang.org/course/detail/(\d+)-(\d+)`)
	if matches == nil || len(matches) < 3 {
		return nil, extractors.ErrURLParseFailed
	}

	heanders := map[string]string{"Origin": "https://time.geekbang.org", "Content-Type": "application/json", "Referer": url}
	params := strings.NewReader("{\"id\":" + string(matches[2]+"}"))
	res, err := request.Request(http.MethodPost, "https://time.geekbang.org/serv/v1/article", params, heanders)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var data geekData
	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	if data.Code < 0 {
		return nil, errors.New(string(data.Error))
	}

	title := data.Data.Title

	streams := make(map[string]downloader.Stream, len(data.Data.VideoMediaMap))

	for key, media := range data.Data.VideoMediaMap {
		m3u8URLs, err := geekM3u8(media.URL)

		if err != nil {
			return nil, err
		}

		urls := make([]downloader.URL, len(m3u8URLs))
		for index, u := range m3u8URLs {
			urls[index] = downloader.URL{
				URL:  u.URL,
				Size: u.Size,
				Ext:  "ts",
			}
		}

		streams[key] = downloader.Stream{
			URLs:    urls,
			Size:    media.Size,
			Quality: key,
		}
	}

	return []downloader.Data{
		{
			Site:    "极客时间 geekbang.org",
			Title:   title,
			Type:    "video",
			Streams: streams,
			URL:     url,
		},
	}, nil
}
