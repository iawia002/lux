package xvideos

import (
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("xvideos", New())
}

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

type extractor struct{}

// New returns a xvideos extractor.
func New() extractors.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var title string
	desc := utils.MatchOneOf(html, `<title>(.+?)</title>`)
	if len(desc) > 1 {
		title = desc[1]
	} else {
		title = "xvideos"
	}

	streams := make(map[string]*extractors.Stream, len(getSrc(html)))
	for _, src := range getSrc(html) {
		size, err := request.Size(src.url, url)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		urlData := &extractors.Part{
			URL:  src.url,
			Size: size,
			Ext:  "mp4",
		}
		streams[src.quality] = &extractors.Stream{
			Parts:   []*extractors.Part{urlData},
			Size:    size,
			Quality: src.quality,
		}
	}
	return []*extractors.Data{
		{
			Site:    "XVIDEOS xvideos.com",
			Title:   title,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
