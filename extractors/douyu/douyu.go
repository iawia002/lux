package douyu

import (
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("douyu", New())
}

type douyuData struct {
	Error int `json:"error"`
	Data  struct {
		VideoURL string `json:"video_url"`
	} `json:"data"`
}

type douyuURLInfo struct {
	URL  string
	Size int64
}

func douyuM3u8(url string) ([]douyuURLInfo, int64, error) {
	var (
		data            []douyuURLInfo
		temp            douyuURLInfo
		size, totalSize int64
		err             error
	)
	urls, err := utils.M3u8URLs(url)
	if err != nil {
		return nil, 0, err
	}
	for _, u := range urls {
		size, err = request.Size(u, url)
		if err != nil {
			return nil, 0, err
		}
		totalSize += size
		temp = douyuURLInfo{
			URL:  u,
			Size: size,
		}
		data = append(data, temp)
	}
	return data, totalSize, nil
}

type extractor struct{}

// New returns a douyu extractor.
func New() extractors.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	var err error
	liveVid := utils.MatchOneOf(url, `https?://www.douyu.com/(\S+)`)
	if liveVid != nil {
		return nil, errors.New("暂不支持斗鱼直播")
	}

	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	titles := utils.MatchOneOf(html, `<title>(.*?)</title>`)
	if titles == nil || len(titles) < 2 {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}
	title := titles[1]

	vids := utils.MatchOneOf(url, `https?://v.douyu.com/show/(\S+)`)
	if vids == nil || len(vids) < 2 {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}
	vid := vids[1]

	dataString, err := request.Get("http://vmobile.douyu.com/video/getInfo?vid="+vid, url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	dataDict := new(douyuData)
	if err := json.Unmarshal([]byte(dataString), dataDict); err != nil {
		return nil, errors.WithStack(err)
	}

	m3u8URLs, totalSize, err := douyuM3u8(dataDict.Data.VideoURL)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	urls := make([]*extractors.Part, len(m3u8URLs))
	for index, u := range m3u8URLs {
		urls[index] = &extractors.Part{
			URL:  u.URL,
			Size: u.Size,
			Ext:  "ts",
		}
	}

	streams := map[string]*extractors.Stream{
		"default": {
			Parts: urls,
			Size:  totalSize,
		},
	}
	return []*extractors.Data{
		{
			Site:    "斗鱼 douyu.com",
			Title:   title,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
