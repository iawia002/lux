package vk

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/iawia002/lux/config"
	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("vk", New())
}

var qualityNames = map[int]string{
	0: "Highest",
	1: "High",
	2: "Medium",
	3: "Low",
	4: "Lowest",
	5: "Legacy",
}

type extractor struct{}

func New() extractors.Extractor {
	return &extractor{}
}

func (e extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	// If url comes from feed or search, its id stored in url parameter.
	// We need to convert it to direct link to make it work with m.vk.com.
	if strings.Contains(url, "z=") {
		split := strings.Split(url, "z=")
		url = split[len(split)-1]
		url = strings.Split(url, "%2F")[0]
	}

	// Convert url to mobile version.
	split := strings.Split(url, "vk.com")
	url = split[len(split)-1]
	if url[0] == '/' {
		url = url[1:]
	}
	url = "https://m.vk.com/" + url

	// Set custom cookies required to download high-res video.
	config.FakeHeaders["Cookie"] += "remixlang=0; remixaudio_show_alert_today=0; remixff=0; remixmdevice=1920/1080/1/!!-!!!!!!"

	// Get html.
	html, err := request.Get(url, url, config.FakeHeaders)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Get video title.
	titles := utils.MatchOneOf(html, `<h1 class="VideoPageInfoRow__title">(.*)</h1>`)
	if titles == nil || len(titles) < 2 {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}
	title := titles[1]

	// Get video urls.
	sources := utils.MatchAll(html, `<source(.*?)/>`)
	srcs := make([]string, len(sources))
	j := 0
	for i := range sources {
		srcs[j] = utils.MatchOneOf(sources[i][1], `src="(.*?)"`)[1]
		srcs[j] = strings.Replace(srcs[j], "&amp;", "&", -1)
		// Some videos have some technical preview on domain vk.com.
		// We need to remove it.
		if strings.Contains(srcs[j], "vk.com") {
			srcs = append(srcs[:j], srcs[j+1:]...)
		} else {
			j++
		}
	}

	// Create download streams.
	streams := make(map[string]*extractors.Stream)
	for i := range srcs {
		size, err := request.Size(srcs[i], "m.vk.vom")
		if err != nil {
			return nil, errors.WithStack(err)
		}
		urlData := &extractors.Part{
			URL:  srcs[i],
			Size: size,
			Ext:  "mp4",
		}
		streams[qualityNames[i]] = &extractors.Stream{
			Parts:   []*extractors.Part{urlData},
			Size:    size,
			Quality: qualityNames[i],
		}
	}

	return []*extractors.Data{
		{
			Site:    "VK vk.com",
			Title:   title,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
