package haokan

import (
	"strings"

	"github.com/iawia002/annie/extractors/types"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type extractor struct{}

// New returns a haokan extractor.
func New() types.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option types.Options) ([]*types.Data, error) {
	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, err
	}

	titles := utils.MatchOneOf(html, `property="og:title"\s+content="(.+?)"`)
	if titles == nil || len(titles) < 2 {
		return nil, types.ErrURLParseFailed
	}
	title := titles[1]

	// 之前的好看网页中，视频地址是放在 video 标签下
	urls := utils.MatchOneOf(html, `<video\s*class="video"\s*src="?(.+?)"?\s*>`)

	if urls == nil || len(urls) < 2 {
		// fallbak: 新的好看网页中，视频地址在 json 数据里
		urls = utils.MatchOneOf(html, `"playurl":"(http.+?)"`)
	}

	if urls == nil || len(urls) < 2 {
		return nil, types.ErrURLParseFailed
	}

	playurl := strings.Replace(urls[1], `\/`, `/`, -1)

	size, err := request.Size(playurl, url)
	if err != nil {
		return nil, err
	}

	_, ext, err := utils.GetNameAndExt(playurl)
	if err != nil {
		return nil, err
	}

	streams := map[string]*types.Stream{
		"default": {
			Parts: []*types.Part{
				{
					URL:  playurl,
					Size: size,
					Ext:  ext,
				},
			},
			Size: size,
		},
	}

	return []*types.Data{
		{
			Site:    "好看视频 haokan.baidu.com",
			Title:   title,
			Type:    types.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
