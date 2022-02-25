package acfun

import (
	"fmt"
	"net/url"
	"regexp"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/parser"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("acfun", New())
}

const (
	bangumiDataPattern = "window.pageInfo = window.bangumiData = (.*);"
	bangumiListPattern = "window.bangumiList = (.*);"

	bangumiHTMLURL = "https://www.acfun.cn/bangumi/aa%d_36188_%d"

	referer = "https://www.acfun.cn"
)

type extractor struct{}

// New returns a new acfun bangumi extractor
func New() extractors.Extractor {
	return &extractor{}
}

// Extract ...
func (e *extractor) Extract(URL string, option extractors.Options) ([]*extractors.Data, error) {
	html, err := request.GetByte(URL, referer, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	epDatas := make([]*episodeData, 0)

	if option.Playlist {
		list, err := resolvingEpisodes(html)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		items := utils.NeedDownloadList(option.Items, option.ItemStart, option.ItemEnd, len(list.Episodes))

		for _, item := range items {
			epDatas = append(epDatas, list.Episodes[item-1])
		}
	} else {
		bgData, _, err := resolvingData(html)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		epDatas = append(epDatas, &bgData.episodeData)
	}

	datas := make([]*extractors.Data, 0)

	wgp := utils.NewWaitGroupPool(option.ThreadNumber)
	for _, epData := range epDatas {
		t := epData
		wgp.Add()
		go func() {
			defer wgp.Done()
			datas = append(datas, extractBangumi(concatURL(t)))
		}()
	}
	wgp.Wait()
	return datas, nil
}

func concatURL(epData *episodeData) string {
	return fmt.Sprintf(bangumiHTMLURL, epData.BangumiID, epData.ItemID)
}

func extractBangumi(URL string) *extractors.Data {
	var err error
	html, err := request.GetByte(URL, referer, nil)
	if err != nil {
		return extractors.EmptyData(URL, err)
	}

	_, vInfo, err := resolvingData(html)
	if err != nil {
		return extractors.EmptyData(URL, err)
	}

	streams := make(map[string]*extractors.Stream)

	for _, stm := range vInfo.AdaptationSet[0].Streams {
		m3u8URL, err := url.Parse(stm.URL)
		if err != nil {
			return extractors.EmptyData(URL, err)
		}

		urls, err := utils.M3u8URLs(m3u8URL.String())
		if err != nil {
			_, err = url.Parse(stm.URL)
			if err != nil {
				return extractors.EmptyData(URL, err)
			}

			urls, err = utils.M3u8URLs(stm.BackURL)
			if err != nil {
				return extractors.EmptyData(URL, err)
			}
		}

		// There is no size information in the m3u8 file and the calculation will take too much time, just ignore it.
		parts := make([]*extractors.Part, 0)
		for _, u := range urls {
			parts = append(parts, &extractors.Part{
				URL: u,
				Ext: "ts",
			})
		}
		streams[stm.QualityLabel] = &extractors.Stream{
			ID:      stm.QualityType,
			Parts:   parts,
			Quality: stm.QualityType,
			NeedMux: false,
		}
	}

	doc, err := parser.GetDoc(string(html))
	if err != nil {
		return extractors.EmptyData(URL, err)
	}
	data := &extractors.Data{
		Site:    "AcFun acfun.cn",
		Title:   parser.Title(doc),
		Type:    extractors.DataTypeVideo,
		Streams: streams,
		URL:     URL,
	}
	return data
}

func resolvingData(html []byte) (*bangumiData, *videoInfo, error) {
	bgData := &bangumiData{}
	vInfo := &videoInfo{}

	pattern, _ := regexp.Compile(bangumiDataPattern)

	groups := pattern.FindSubmatch(html)
	err := jsoniter.Unmarshal(groups[1], bgData)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	err = jsoniter.UnmarshalFromString(bgData.CurrentVideoInfo.KsPlayJSON, vInfo)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	return bgData, vInfo, nil
}

func resolvingEpisodes(html []byte) (*episodeList, error) {
	list := &episodeList{}
	pattern, _ := regexp.Compile(bangumiListPattern)

	groups := pattern.FindSubmatch(html)
	err := jsoniter.Unmarshal(groups[1], list)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return list, nil
}
