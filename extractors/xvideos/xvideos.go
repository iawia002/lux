package xvideos

import (
	"fmt"
	"strings"
	"sync"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

const (
	lowFlag      = "html5player.setVideoUrlLow('"
	lowFinalFlag = `');
	    html5player.setVideoUrlHigh(`
	highFlag      = "html5player.setVideoUrlHigh('"
	highFinalFlag = `');
	    html5player.setVideoHLS(`
	qualityLow  = "low"
	qualityHigh = "high"
)

var (
	lowFlagLength  = len(lowFlag)
	highFlagLength = len(highFlag)
)

type src struct {
	url     string
	quality string
}

func getSrc(html string) []*src {
	var wg sync.WaitGroup
	wg.Add(4)

	startIndexLow := 0
	go func() {
		startIndexLow = strings.Index(html, lowFlag)
		startIndexLow += lowFlagLength
		wg.Done()
	}()
	endIndexLow := 0
	go func() {
		endIndexLow = strings.Index(html, lowFinalFlag)
		wg.Done()
	}()

	startIndexHigh := 0
	go func() {
		startIndexHigh = strings.Index(html, highFlag)
		startIndexHigh += highFlagLength
		wg.Done()
	}()
	endIndexHigh := 0
	go func() {
		endIndexHigh = strings.Index(html, highFinalFlag)
		wg.Done()
	}()
	wg.Wait()

	var srcs []*src
	if startIndexLow != -1 {
		srcs = append(srcs, &src{
			url:     html[startIndexLow:endIndexLow],
			quality: qualityLow,
		})
	}
	if startIndexHigh != -1 {
		srcs = append(srcs, &src{
			url:     html[startIndexHigh:endIndexHigh],
			quality: qualityHigh,
		})
	}
	return srcs
}

// Extract is the main function for extracting data
func Extract(url string) ([]downloader.Data, error) {
	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, err
	}
	var title string
	desc := utils.MatchOneOf(html, `<title>(.+?)</title>`)
	if desc != nil && len(desc) > 1 {
		title = desc[1]
	} else {
		title = "xvideos"
	}

	streams := make(map[string]downloader.Stream)
	for _, src := range getSrc(html) {
		size, err := request.Size(src.url, url)
		if err != nil {
			return nil, err
		}
		urlData := downloader.URL{
			URL:  src.url,
			Size: size,
			Ext:  "mp4",
		}
		streams[src.quality] = downloader.Stream{
			URLs:    []downloader.URL{urlData},
			Size:    size,
			Quality: fmt.Sprintf("%s", src.quality),
		}
	}
	return []downloader.Data{
		{
			Site:    "XVIDEOS xvideos.com",
			Title:   title,
			Type:    "video",
			Streams: streams,
			URL:     url,
		},
	}, nil
}
