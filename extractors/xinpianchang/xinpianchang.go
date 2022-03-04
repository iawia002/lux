package xinpianchang

import (
	"encoding/json"
	"fmt"
	"regexp"

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
	URL     string `json:"url"`
	Size    int64  `json:"size"`
	Ext     string `json:"ext"`
	Quality string `json:"quality"`
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

	r1 := regexp.MustCompile(`vid = \"(.+?)\";`)
	r2 := regexp.MustCompile(`modeServerAppKey = \"(.+?)\";`)

	vid := r1.FindSubmatch([]byte(html))[1]
	appKey := r2.FindSubmatch([]byte(html))[1]

	video_url := fmt.Sprintf("https://mod-api.xinpianchang.com/mod/api/v2/media/%s?appKey=%s", string(vid), string(appKey))
	body, err := request.Get(video_url, url, headers)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	streams := make(map[string]*extractors.Stream)

	var m interface{}
	err = json.Unmarshal([]byte(body), &m)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var title string
	query1, err := gojq.Parse(".data.title")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	iter := query1.Run(m)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return nil, errors.WithStack(err)
		}
		title, _ = v.(string)
	}

	query2, err := gojq.Parse(".data.resource.progressive[] | {quality: .quality, size: .filesize, url: .url}")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	iter2 := query2.Run(m)
	for {
		v, ok := iter2.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return nil, errors.WithStack(err)
		}

		video := Video{}

		jsonbody, err := json.Marshal(v)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		if err := json.Unmarshal(jsonbody, &video); err != nil {
			return nil, errors.WithStack(err)
		}
		video.Ext = "mp4"

		stream := extractors.Stream{
			Size:    video.Size,
			Quality: video.Quality,
			Parts: []*extractors.Part{
				&extractors.Part{
					URL:  video.URL,
					Size: video.Size,
					Ext:  video.Ext,
				},
			},
		}
		streams[video.Quality] = &stream
	}

	return []*extractors.Data{
		{
			Site:    "新片场 xinpianchang.com",
			Title:   title,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
