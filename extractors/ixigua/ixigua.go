package ixigua

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

const (
	referer       = "https://www.ixigua.com"
	defaultCookie = "__ac_nonce=0621792ac001d3194585; __ac_signature=_02B4Z6wo00f01N2VvVQAAIDBW72Y87LFxvzdtLnAAFV4fe; __ac_referer=__ac_blank; ttwid_date=1"
)

type extractor struct{}

// New returns a ixigua extractor.
func New() extractors.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	html, err := request.Get(url, referer, map[string]string{
		"Cookie":     getCookie(option.Cookie),
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:98.0) Gecko/20100101 Firefox/98.0",
	})
	println(html)
	if err != nil {
		return nil, err
	}
	jsonRegexp := utils.MatchOneOf(html, `window\._SSR_HYDRATED_DATA=(.*?)</script>`)
	if jsonRegexp == nil || len(jsonRegexp) < 2 {
		return nil, extractors.ErrURLParseFailed
	}
	jsonStr := strings.ReplaceAll(string(jsonRegexp[1]), ":undefined", ":\"undefined\"")
	var (
		title   string
		streams map[string]*extractors.Stream
	)
	if regexp.MustCompile(`"albumId"`).MatchString(html) {
		var ratedData ssrHydratedDataEpisode
		if err := json.Unmarshal([]byte(jsonStr), &ratedData); err != nil {
			return nil, err
		}
		episodeInfo := ratedData.AnyVideo.GidInformation.PackerData.EpisodeInfo
		title = fmt.Sprintf("%s %s", episodeInfo.Title, episodeInfo.Name)
		if streams, err = getStreamsEpisode(&ratedData); err != nil {
			return nil, err
		}
	} else {
		var ratedData ssrHydratedData
		if err := json.Unmarshal([]byte(jsonStr), &ratedData); err != nil {
			return nil, err
		}
		title = ratedData.AnyVideo.GidInformation.PackerData.Video.Title
		if streams, err = getStreams(&ratedData); err != nil {
			return nil, err
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

func getCookie(c string) string {
	if c != "" {
		return c
	}
	return defaultCookie
}

func getStreams(d *ssrHydratedData) (map[string]*extractors.Stream, error) {
	streams := make(map[string]*extractors.Stream)
	videoList := d.AnyVideo.GidInformation.PackerData.Video.VideoResource.Dash120Fps.DynamicVideo.DynamicVideoList
	audioList := d.AnyVideo.GidInformation.PackerData.Video.VideoResource.Dash120Fps.DynamicVideo.DynamicAudioList
	audioCount := len(audioList)
	if audioCount > 0 {
		audioURL := base64Decode(audioList[audioCount-1].MainURL)
		audioSize, err := request.Size(audioURL, referer)
		audioPart := &extractors.Part{
			URL:  audioURL,
			Size: audioSize,
			Ext:  "mp3",
		}
		if err != nil {
			return nil, err
		}
		for _, i := range videoList {
			if i.MainURL == "" {
				continue
			}
			videoURL := base64Decode(i.MainURL)
			videoSize, err := request.Size(videoURL, referer)
			if err != nil {
				return nil, err
			}
			videoPart := &extractors.Part{
				URL:  videoURL,
				Size: videoSize,
				Ext:  "mp4",
			}
			streams[i.Definition] = &extractors.Stream{
				ID:      i.Definition,
				Quality: i.Definition,
				Parts:   []*extractors.Part{videoPart, audioPart},
				Size:    audioSize + videoSize,
				Ext:     "mp4",
				NeedMux: true,
			}

		}
		return streams, nil
	}
	Normal := d.AnyVideo.GidInformation.PackerData.Video.VideoResource.Normal.VideoList
	NormalList := []video{Normal.Video1, Normal.Video2, Normal.Video3, Normal.Video4}
	for _, i := range NormalList {
		if i.MainURL == "" {
			continue
		}
		videoURL := base64Decode(i.MainURL)
		videoSize, err := request.Size(videoURL, referer)
		if err != nil {
			return nil, err
		}
		videoPart := &extractors.Part{
			URL:  videoURL,
			Size: videoSize,
			Ext:  "mp4",
		}
		streams[i.Definition] = &extractors.Stream{
			ID:      i.Definition,
			Quality: i.Definition,
			Parts:   []*extractors.Part{videoPart},
			Size:    videoSize,
			Ext:     "mp4",
		}
	}
	return streams, nil
}

func getStreamsEpisode(d *ssrHydratedDataEpisode) (map[string]*extractors.Stream, error) {
	streams := make(map[string]*extractors.Stream)
	Normal := d.AnyVideo.GidInformation.PackerData.VideoResource.Normal.VideoList
	NormalList := []video{Normal.Video1, Normal.Video2, Normal.Video3, Normal.Video4}
	for _, i := range NormalList {
		if i.MainURL == "" {
			continue
		}
		videoURL := base64Decode(i.MainURL)
		videoSize, err := request.Size(videoURL, referer)
		if err != nil {
			return nil, err
		}
		videoPart := &extractors.Part{
			URL:  videoURL,
			Size: videoSize,
			Ext:  "mp4",
		}
		streams[i.Definition] = &extractors.Stream{
			ID:      i.Definition,
			Quality: i.Definition,
			Parts:   []*extractors.Part{videoPart},
			Size:    videoSize,
		}
	}
	return streams, nil
}

func base64Decode(t string) string {
	d, _ := base64.StdEncoding.DecodeString(t)
	return string(d)
}
