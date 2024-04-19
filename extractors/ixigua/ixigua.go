package ixigua

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("ixigua", New())
	extractors.Register("toutiao", New())
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

// New returns a ixigua extractor.
func New() extractors.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	headers := map[string]string{
		"User-Agent": browser.Chrome(),
		"Cookie":     option.Cookie,
	}

	// ixigua 有三种格式的 URL
	// 格式一 https://www.ixigua.com/7053389963487871502
	// 格式二 https://v.ixigua.com/RedcbWM/
	// 格式三 https://m.toutiao.com/is/dtj1pND/
	// 格式二会跳转到格式一
	// 格式三会跳转到 https://www.toutiao.com/a7053389963487871502

	var finalURL string
	if strings.HasPrefix(url, "https://www.ixigua.com/") {
		finalURL = url
	}

	if strings.HasPrefix(url, "https://v.ixigua.com/") || strings.HasPrefix(url, "https://m.toutiao.com/") {
		resp, err := http.Get(url)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		defer resp.Body.Close() // nolint
		// follow redirects, https://stackoverflow.com/a/16785343
		finalURL = resp.Request.URL.String()
	}

	finalURL = strings.ReplaceAll(finalURL, "https://www.toutiao.com/video/", "https://www.ixigua.com/")

	r := regexp.MustCompile(`(ixigua.com/)(\w+)?`)
	id := r.FindSubmatch([]byte(finalURL))[2]
	url2 := fmt.Sprintf("https://www.ixigua.com/%s", string(id))

	body, err := request.Get(url2, url, headers)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	videoListJson := utils.MatchOneOf(body, `window._SSR_HYDRATED_DATA=(\{.*?\})\<\/script\>`)
	if videoListJson == nil || len(videoListJson) != 2 {
		return nil, errors.WithStack(extractors.ErrBodyParseFailed)
	}

	videoUrl := videoListJson[1]
	videoUrl = strings.Replace(videoUrl, ":undefined", ":\"undefined\"", -1)

	var data xiguanData
	if err = json.Unmarshal([]byte(videoUrl), &data); err != nil {
		return nil, errors.WithStack(err)
	}

	title := data.AnyVideo.GidInformation.PackerData.Video.Title
	videoList := data.AnyVideo.GidInformation.PackerData.Video.VideoResource.Normal.VideoList

	streams := make(map[string]*extractors.Stream)
	for _, v := range videoList {
		streams[v.Definition] = &extractors.Stream{
			Quality: v.Definition,
			Parts: []*extractors.Part{
				{
					URL:  base64Decode(v.MainUrl),
					Size: v.Size,
					Ext:  v.Vtype,
				},
			},
		}
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
