package dailymotion

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

const (
	qualityAuto = "auto"
	quality144  = "144p"
	quality240  = "240p"
	quality380  = "380p"
	quality480  = "480p"
	quality720  = "720p"
)

type qualities struct {
	Auto []*src `json:"auto"`
	Q1   []*src `json:"144"`
	Q2   []*src `json:"240"`
	Q3   []*src `json:"380"`
	Q4   []*src `json:"480"`
	Q5   []*src `json:"720"`
}

type src struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type htmlContext struct {
	Metadata struct {
		Qualities *qualities `json:"qualities"`
	} `json:"metadata"`
}

func getSrc(html string) (*qualities, error) {
	htmlCtx := &htmlContext{}
	if jsonSrc := utils.MatchOneOf(html, `var config = (.+?);`); len(jsonSrc) > 1 && jsonSrc[1] != "" {
		if err := json.Unmarshal([]byte(jsonSrc[1]), htmlCtx); err != nil {
			return nil, err
		}
		return htmlCtx.Metadata.Qualities, nil
	}
	return nil, errors.New("parse html fail")
}

func handleM3u8U(url string) ([]downloader.URL, int64, error) {
	urls, err := utils.M3u8URLs(url)
	if err != nil {
		return nil, 0, err
	}
	var totalSize int64
	data := make([]downloader.URL, len(urls))
	for i, u := range urls {
		size, err := request.Size(u, url)
		if err != nil {
			return nil, 0, err
		}
		data[i] = downloader.URL{
			URL:  u,
			Size: size,
			Ext:  "ts",
		}
		totalSize += size
	}
	return data, totalSize, nil
}

func handleMP4(srcURL, ref string) (downloader.URL, error) {
	size, err := request.Size(srcURL, ref)
	if err != nil {
		return downloader.URL{}, err
	}
	return downloader.URL{
		URL:  srcURL,
		Size: size,
		Ext:  "mp4",
	}, nil
}

func handle(srcs []*src, streams map[string]downloader.Stream, quality, refURL string) error {
	for _, src := range srcs {
		if src.Type == "application/x-mpegURL" {
			drs, totalSize, err := handleM3u8U(src.URL)
			if err != nil {
				return err
			}
			streams[quality] = downloader.Stream{
				URLs:    drs,
				Size:    totalSize,
				Quality: quality,
			}
			continue
		}
		if src.Type == "video/mp4" {
			dr, err := handleMP4(src.URL, refURL)
			if err != nil {
				return err
			}
			streams[quality] = downloader.Stream{
				URLs:    []downloader.URL{dr},
				Size:    dr.Size,
				Quality: quality,
			}
		}
	}
	return nil
}

func prepareEmbedURL(url string) string {
	if !strings.Contains(url, "https://www.dailymotion.com/embed/") {
		newIDs := strings.Split(url, "/")
		return "https://www.dailymotion.com/embed/video/" + newIDs[len(newIDs)-1]
	}
	return url
}

// Extract is the main function for extracting data
func Extract(url string) ([]downloader.Data, error) {
	url = prepareEmbedURL(url)
	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, err
	}
	var title string
	if desc := utils.MatchOneOf(html, `<title>(.+?)</title>`); desc != nil {
		title = desc[1]
	} else {
		title = "dailymotion"
	}
	title = strings.Replace(title, "Dailymotion Video Player - ", "", 1)
	streams := make(map[string]downloader.Stream)
	qts, err := getSrc(html)
	if err != nil {
		return nil, err
	}
	if err = handle(qts.Auto, streams, qualityAuto, url); err != nil {
		return nil, err
	}
	if err = handle(qts.Q1, streams, quality144, url); err != nil {
		return nil, err
	}
	if err = handle(qts.Q2, streams, quality240, url); err != nil {
		return nil, err
	}
	if err = handle(qts.Q3, streams, quality380, url); err != nil {
		return nil, err
	}
	if err = handle(qts.Q4, streams, quality480, url); err != nil {
		return nil, err
	}
	if err = handle(qts.Q5, streams, quality720, url); err != nil {
		return nil, err
	}
	return []downloader.Data{
		{
			Site:    "Dailymotion dailymotion.com",
			Title:   title,
			Type:    "video",
			Streams: streams,
			URL:     url,
		},
	}, nil
}
