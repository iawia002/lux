package haokan

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("haokan", New())
}

type extractor struct{}

// New returns a haokan extractor.
func New() extractors.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	titles := utils.MatchOneOf(html, `property="og:title"\s+content="(.+?)"`)
	if titles == nil || len(titles) < 2 {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}
	title := titles[1]

	// 之前的好看网页中，视频地址是放在 video 标签下
	urls := utils.MatchOneOf(html, `<video\s*class="video"\s*src="?(.+?)"?\s*>`)

	if urls == nil || len(urls) < 2 {
		// fallbak: 新的好看网页中，视频地址在 json 数据里
		urls = utils.MatchOneOf(html, `"playurl":"(http.+?)"`)
	}

	if urls == nil || len(urls) < 2 {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}

	playurl := strings.Replace(urls[1], `\/`, `/`, -1)

	size, err := request.Size(playurl, url)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	_, ext, err := utils.GetNameAndExt(playurl)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	streams := map[string]*extractors.Stream{
		"default": {
			Parts: []*extractors.Part{
				{
					URL:  playurl,
					Size: size,
					Ext:  ext,
				},
			},
			Size: size,
		},
	}

	return []*extractors.Data{
		{
			Site:    "好看视频 haokan.baidu.com",
			Title:   title,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
