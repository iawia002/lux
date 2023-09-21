package extractors

import (
	"net/url"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/wujiu2020/lux/extractors/bilibili"
	"github.com/wujiu2020/lux/extractors/cctv"
	"github.com/wujiu2020/lux/extractors/douyin"
	"github.com/wujiu2020/lux/extractors/iqiyi"
	"github.com/wujiu2020/lux/extractors/proto"
	"github.com/wujiu2020/lux/extractors/qq"
	"github.com/wujiu2020/lux/utils"
)

func init() {
	douyin := douyin.New()
	Register("douyin", douyin)
	Register("iesdouyin", douyin)
	Register("douyin", bilibili.New())
	Register("iqiyi", iqiyi.New(iqiyi.SiteTypeIqiyi))
	Register("iq", iqiyi.New(iqiyi.SiteTypeIQ))
	Register("qq", qq.New())
	Register("cctv", cctv.New())
}

var lock sync.RWMutex
var extractorMap = make(map[string]proto.Extractor)

// Register registers an Extractor.
func Register(domain string, e proto.Extractor) {
	lock.Lock()
	extractorMap[domain] = e
	lock.Unlock()
}

// Extract is the main function to extract the data.
func Extract(u string) (*proto.Data, error) {
	u = strings.TrimSpace(u)
	var domain string

	bilibiliShortLink := utils.MatchOneOf(u, `^(av|BV|ep)\w+`)
	if len(bilibiliShortLink) > 1 {
		bilibiliURL := map[string]string{
			"av": "https://www.bilibili.com/video/",
			"BV": "https://www.bilibili.com/video/",
			"ep": "https://www.bilibili.com/bangumi/play/",
		}
		domain = "bilibili"
		u = bilibiliURL[bilibiliShortLink[1]] + u
	} else {
		u, err := url.ParseRequestURI(u)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		domain = utils.Domain(u.Host)
	}
	extractor := extractorMap[domain]
	if extractor == nil {
		extractor = extractorMap[""]
	}
	videos, err := extractor.Extract(u)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return videos, nil
}
