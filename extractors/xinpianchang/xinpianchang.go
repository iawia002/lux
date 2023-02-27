package xinpianchang

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/itchyny/gojq"
	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
)

func init() {
	extractors.Register("xinpianchang", New())
}

type extractor struct{}

type Video struct {
	Title     string `json:"title"`
	Qualities []struct {
		Quality string `json:"quality"`
		Size    int64  `json:"size"`
		URL     string `json:"url"`
		Ext     string `json:"ext"`
	} `json:"qualities"`
}

// New returns a xinpianchang extractor.
func New() extractors.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:98.0) Gecko/20100101 Firefox/98.0",
	}

	html, err := request.Get(url, url, headers)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	r1 := regexp.MustCompile(`vid = "(.+?)";`)
	r2 := regexp.MustCompile(`modeServerAppKey = "(.+?)";`)

	vid := r1.FindSubmatch([]byte(html))[1]
	appKey := r2.FindSubmatch([]byte(html))[1]

	video_url := fmt.Sprintf("https://mod-api.xinpianchang.com/mod/api/v2/media/%s?appKey=%s", string(vid), string(appKey))
	body, err := request.Get(video_url, url, headers)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var m interface{}
	err = json.Unmarshal([]byte(body), &m)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	query, err := gojq.Parse("{title: .data.title} + {qualities: [(.data.resource.progressive[] | {quality: .quality, size: .filesize, url: .url, ext: .mime})]}")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	iter := query.Run(m)
	video := Video{}

	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return nil, errors.WithStack(err)
		}

		jsonbody, err := json.Marshal(v)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		if err := json.Unmarshal(jsonbody, &video); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	streams := make(map[string]*extractors.Stream)
	for _, quality := range video.Qualities {
		streams[quality.Quality] = &extractors.Stream{
			Size:    quality.Size,
			Quality: quality.Quality,
			Parts: []*extractors.Part{
				{
					URL:  quality.URL,
					Size: quality.Size,
					Ext:  strings.Split(quality.Ext, "/")[1],
				},
			},
		}
	}

	return []*extractors.Data{
		{
			Site:    "新片场 xinpianchang.com",
			Title:   video.Title,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
