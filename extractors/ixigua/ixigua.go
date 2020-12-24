package ixigua

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/iawia002/annie/extractors/types"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

const (
	referer   = "https://www.ixigua.com"
	defCookie = "xiguavideopcwebid=6872983880459503118; xiguavideopcwebid.sig=B4DvNwwGGQ-hDxYcJo5FfbMEIn4; _ga=GA1.2.572711536.1600241266; MONITOR_WEB_ID=bfe0e43a-e004-400e-8040-81677a199a22; ttwid=1%7CPWHvUSGTtsxK0WUzkuq7SxJtT7L3WHRvJeSGG5WZjiw%7C1604995289%7Cec6a591ac986362929a9be173d65df8f3551269fff0694d34a5e57a33cd287eb; ixigua-a-s=1; Hm_lvt_db8ae92f7b33b6596893cdf8c004a1a2=1608261601; _gid=GA1.2.1203395873.1608261601; Hm_lpvt_db8ae92f7b33b6596893cdf8c004a1a2=1608262109"
)

type extractor struct{}

// New returns a youtube extractor.
func New() types.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option types.Options) ([]*types.Data, error) {
	html, err := request.Get(url, referer, map[string]string{
		"cookie": getCookie(option.Cookie),
	})
	if err != nil {
		return nil, err
	}
	jsonRegexp := utils.MatchOneOf(html, `window\._SSR_HYDRATED_DATA=(.*?)</script>`)
	if jsonRegexp == nil || len(jsonRegexp) < 2 {
		return nil, types.ErrURLParseFailed
	}
	jsonStr := strings.ReplaceAll(string(jsonRegexp[1]), ":undefined", ":\"undefined\"")
	var (
		title   string
		streams map[string]*types.Stream
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
	return []*types.Data{
		{
			Site:    "西瓜视频 ixigua.com",
			Title:   title,
			Type:    types.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil

}

func getCookie(c string) string {
	if c != "" {
		return c
	}
	return defCookie
}

func getStreams(d *ssrHydratedData) (map[string]*types.Stream, error) {
	streams := make(map[string]*types.Stream)
	videoList := d.AnyVideo.GidInformation.PackerData.Video.VideoResource.Dash120Fps.DynamicVideo.DynamicVideoList
	audioList := d.AnyVideo.GidInformation.PackerData.Video.VideoResource.Dash120Fps.DynamicVideo.DynamicAudioList
	audioCount := len(audioList)
	if audioCount > 0 {
		audioURL := base64Decode(audioList[audioCount-1].MainURL)
		audioSize, err := request.Size(audioURL, referer)
		audioPart := &types.Part{
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
			videoPart := &types.Part{
				URL:  videoURL,
				Size: videoSize,
				Ext:  "mp4",
			}
			streams[i.Definition] = &types.Stream{
				ID:      i.Definition,
				Quality: i.Definition,
				Parts:   []*types.Part{videoPart, audioPart},
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
		videoPart := &types.Part{
			URL:  videoURL,
			Size: videoSize,
			Ext:  "mp4",
		}
		streams[i.Definition] = &types.Stream{
			ID:      i.Definition,
			Quality: i.Definition,
			Parts:   []*types.Part{videoPart},
			Size:    videoSize,
			Ext:     "mp4",
		}
	}
	return streams, nil
}

func getStreamsEpisode(d *ssrHydratedDataEpisode) (map[string]*types.Stream, error) {
	streams := make(map[string]*types.Stream)
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
		videoPart := &types.Part{
			URL:  videoURL,
			Size: videoSize,
			Ext:  "mp4",
		}
		streams[i.Definition] = &types.Stream{
			ID:      i.Definition,
			Quality: i.Definition,
			Parts:   []*types.Part{videoPart},
			Size:    videoSize,
		}
	}
	return streams, nil
}

func base64Decode(t string) string {
	d, _ := base64.StdEncoding.DecodeString(t)
	return string(d)
}
