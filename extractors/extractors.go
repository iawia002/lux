package extractors

import (
	"net/url"
	"strings"

	"github.com/iawia002/lux/extractors/acfun"
	"github.com/iawia002/lux/extractors/bcy"
	"github.com/iawia002/lux/extractors/bilibili"
	"github.com/iawia002/lux/extractors/douyin"
	"github.com/iawia002/lux/extractors/douyu"
	"github.com/iawia002/lux/extractors/eporner"
	"github.com/iawia002/lux/extractors/facebook"
	"github.com/iawia002/lux/extractors/geekbang"
	"github.com/iawia002/lux/extractors/haokan"
	"github.com/iawia002/lux/extractors/hupu"
	"github.com/iawia002/lux/extractors/instagram"
	"github.com/iawia002/lux/extractors/iqiyi"
	"github.com/iawia002/lux/extractors/mgtv"
	"github.com/iawia002/lux/extractors/miaopai"
	"github.com/iawia002/lux/extractors/netease"
	"github.com/iawia002/lux/extractors/pixivision"
	"github.com/iawia002/lux/extractors/pornhub"
	"github.com/iawia002/lux/extractors/qq"
	"github.com/iawia002/lux/extractors/streamtape"
	"github.com/iawia002/lux/extractors/tangdou"
	"github.com/iawia002/lux/extractors/tiktok"
	"github.com/iawia002/lux/extractors/tumblr"
	"github.com/iawia002/lux/extractors/twitter"
	"github.com/iawia002/lux/extractors/types"
	"github.com/iawia002/lux/extractors/udn"
	"github.com/iawia002/lux/extractors/universal"
	"github.com/iawia002/lux/extractors/vimeo"
	"github.com/iawia002/lux/extractors/weibo"
	"github.com/iawia002/lux/extractors/xvideos"
	"github.com/iawia002/lux/extractors/yinyuetai"
	"github.com/iawia002/lux/extractors/youku"
	"github.com/iawia002/lux/extractors/youtube"
	"github.com/iawia002/lux/utils"
)

var extractorMap map[string]types.Extractor

func init() {
	douyinExtractor := douyin.New()
	youtubeExtractor := youtube.New()
	stExtractor := streamtape.New()

	extractorMap = map[string]types.Extractor{
		"": universal.New(), // universal extractor

		"douyin":     douyinExtractor,
		"iesdouyin":  douyinExtractor,
		"bilibili":   bilibili.New(),
		"bcy":        bcy.New(),
		"pixivision": pixivision.New(),
		"youku":      youku.New(),
		"youtube":    youtubeExtractor,
		"youtu":      youtubeExtractor, // youtu.be
		"iqiyi":      iqiyi.New(iqiyi.SiteTypeIqiyi),
		"iq":         iqiyi.New(iqiyi.SiteTypeIQ),
		"mgtv":       mgtv.New(),
		"tangdou":    tangdou.New(),
		"tumblr":     tumblr.New(),
		"vimeo":      vimeo.New(),
		"facebook":   facebook.New(),
		"douyu":      douyu.New(),
		"miaopai":    miaopai.New(),
		"163":        netease.New(),
		"weibo":      weibo.New(),
		"instagram":  instagram.New(),
		"twitter":    twitter.New(),
		"qq":         qq.New(),
		"yinyuetai":  yinyuetai.New(),
		"geekbang":   geekbang.New(),
		"pornhub":    pornhub.New(),
		"xvideos":    xvideos.New(),
		"udn":        udn.New(),
		"tiktok":     tiktok.New(),
		"haokan":     haokan.New(),
		"acfun":      acfun.New(),
		"eporner":    eporner.New(),
		"streamtape": stExtractor,
		"streamta":   stExtractor, // streamta.pe
		"hupu":       hupu.New(),
	}
}

// Extract is the main function to extract the data.
func Extract(u string, option types.Options) ([]*types.Data, error) {
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
			return nil, err
		}
		if u.Host == "haokan.baidu.com" {
			domain = "haokan"
		} else {
			domain = utils.Domain(u.Host)
		}
	}
	extractor := extractorMap[domain]
	if extractor == nil {
		extractor = extractorMap[""]
	}
	videos, err := extractor.Extract(u, option)
	if err != nil {
		return nil, err
	}
	for _, v := range videos {
		v.FillUpStreamsData()
	}
	return videos, nil
}
