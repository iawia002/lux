package extractors

import (
	"net/url"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/wujiu2020/lux/extractors/cctv"
	"github.com/wujiu2020/lux/extractors/proto"
	"github.com/wujiu2020/lux/utils"
	// "github.com/wujiu2020/lux/extractors/douyin"
	// "github.com/wujiu2020/lux/extractors/iqiyi"
	// "github.com/wujiu2020/lux/extractors/qq"
)

func init() {
	// douyin := douyin.New()
	// Register("douyin", douyin)
	// Register("iesdouyin", douyin)
	// Register("iqiyi", iqiyi.New(iqiyi.SiteTypeIqiyi))
	// Register("iq", iqiyi.New(iqiyi.SiteTypeIQ))
	// Register("qq", qq.New())
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
func Extract(u, quality string) (*proto.Data, error) {
	u = strings.TrimSpace(u)
	var domain string

	uri, err := url.ParseRequestURI(u)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	domain = utils.Domain(uri.Host)
	extractor := extractorMap[domain]
	if extractor == nil {
		extractor = extractorMap[""]
	}
	videos, err := extractor.Extract(u)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return videos.TransformData(u, quality), nil
}
