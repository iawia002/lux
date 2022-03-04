package ixigua

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/itchyny/gojq"
	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
)

func init() {
	extractors.Register("ixigua", New())
}

type extractor struct{}

type Video struct {
	URL     string `json:"url"`
	Size    int64  `json:"size"`
	Ext     string `json:"ext"`
	Quality string `json:"quality"`
}

// New returns a extractor.
func New() extractors.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	headers := map[string]string{
		"User-Agent":   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:98.0) Gecko/20100101 Firefox/98.0",
		"Content-Type": "application/json",
	}

	streams := make(map[string]*extractors.Stream)

	r := regexp.MustCompile(`(ixigua.com/)(\w+)?`)
	id := r.FindSubmatch([]byte(url))[2]
	url2 := fmt.Sprintf("https://www.ixigua.com/api/public/videov2/brief/details?group_id=%s", string(id))

	body, err := request.Get(url2, url, headers)
	if err != nil {
		return nil, errors.WithStack(err)
	}

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

	query2, err := gojq.Parse(".data.videoResource.normal.video_list | .[] | {url: .main_url, size: .size, ext: .vtype, quality: .definition}")
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
		video.URL = base64Decode(video.URL)

		stream := extractors.Stream{
			Quality: video.Quality,
			Parts: []*extractors.Part{
				&extractors.Part{
					URL:  video.URL,
					Size: video.Size,
					Ext:  video.Ext,
				},
			},
			NeedMux: false,
		}
		streams[video.Quality] = &stream
	}

	return []*extractors.Data{
		{
			Site:    "西瓜视频 ixigua.com",
			Title:   title,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil

}

func base64Decode(t string) string {
	d, _ := base64.StdEncoding.DecodeString(t)
	return string(d)
}
