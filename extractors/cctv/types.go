package cctv

import (
	"strconv"
	"strings"

	"github.com/wujiu2020/lux/extractors/proto"
)

const (
	QualitySd1  = "h264418000"  //流畅   270
	QualityHd1  = "h264818000"  //标清   480
	QualityShd1 = "h2641200000" //高清   720
	QualityFhd1 = "h2642000000" //超光   1080
)

const (
	QualitySd2  = "h264_418"  //流畅   270
	QualityHd2  = "h264_818"  //标清   480
	QualityShd2 = "h264_1072" //高清   720
	QualityFhd2 = "h264_1872" //超光   1080
)

var qualityMap1 = map[string]string{
	"270P":  QualitySd1,
	"480P":  QualityHd1,
	"720P":  QualityShd1,
	"1080P": QualityFhd1,
}

var qualityMap2 = map[string]string{
	"270P":  QualitySd2,
	"480P":  QualityHd2,
	"720P":  QualityShd2,
	"1080P": QualityFhd2,
}

const (
	Api = "https://player-api.yangshipin.cn/v1/player/get_video_info"
)

type VideoInfoRes struct {
	Ack         string `json:"ack"`
	Status      string `json:"status"`
	Tag         string `json:"tag"`
	Title       string `json:"title"`
	PlayChannel string `json:"play_channel"`
	Produce     string `json:"produce"`
	EditerName  string `json:"editer_name"`
	ProduceID   string `json:"produce_id"`
	Column      string `json:"column"`
	FPgmtime    string `json:"f_pgmtime"`
	CdnInfo     struct {
		CdnVip  string `json:"cdn_vip"`
		CdnCode string `json:"cdn_code"`
		CdnName string `json:"cdn_name"`
	} `json:"cdn_info"`
	Video struct {
		TotalLength     string    `json:"totalLength"`
		Chapters        []Chapter `json:"chapters"`
		Chapters2       []Chapter `json:"chapters2"`
		Chapters3       []Chapter `json:"chapters3"`
		Chapters4       []Chapter `json:"chapters4"`
		ValidChapterNum int       `json:"validChapterNum"`
		URL             string    `json:"url"`
	} `json:"video"`
	HlsCdnInfo struct {
		CdnVip  string `json:"cdn_vip"`
		CdnCode string `json:"cdn_code"`
		CdnName string `json:"cdn_name"`
	} `json:"hls_cdn_info"`
	HlsURL       string `json:"hls_url"`
	AspErrorCode string `json:"asp_error_code"`
	Manifest     struct {
		AudioMp3    string `json:"audio_mp3"`
		HlsAudioURL string `json:"hls_audio_url"`
		HlsEncURL   string `json:"hls_enc_url"`
		HlsH5EURL   string `json:"hls_h5e_url"`
		HlsEnc2URL  string `json:"hls_enc2_url"`
	} `json:"manifest"`
	ClientSid          string `json:"client_sid"`
	Public             string `json:"public"`
	IsInvalidCopyright string `json:"is_invalid_copyright"`
	IsProtected        string `json:"is_protected"`
	IsFnHot            string `json:"is_fn_hot"`
	IsP2PUse           bool   `json:"is_p2p_use"`
	DefaultStream      string `json:"default_stream"`
	Lc                 struct {
		IspCode     string `json:"isp_code"`
		CityCode    string `json:"city_code"`
		ProviceCode string `json:"provice_code"`
		CountryCode string `json:"country_code"`
		IP          string `json:"ip"`
	} `json:"lc"`
	IsIpadSupport   string `json:"is_ipad_support"`
	Version         string `json:"version"`
	Embed           string `json:"embed"`
	IsFnMultiStream bool   `json:"is_fn_multi_stream"`
}

type Chapter struct {
	Duration string `json:"duration"`
	Image    string `json:"image"`
	URL      string `json:"url"`
}

func (v VideoInfoRes) TransformData(url string, quality string) *proto.Data {
	var data proto.Data
	data.Title = v.Title
	totalDuration, _ := strconv.ParseFloat(v.Video.TotalLength, 64)
	data.Duration = totalDuration
	chaptersList := [][]Chapter{v.Video.Chapters, v.Video.Chapters2, v.Video.Chapters3, v.Video.Chapters4}
	if qualityMap1[quality] == "" && qualityMap2[quality] == "" {
		quality = "270P"
	}
	for _, chapters := range chaptersList {
		var hasStream bool
		stream := proto.Stream{
			Referer:   url,
			Useragent: "",
			Quality:   quality,
		}
		for _, chapter := range chapters {
			if !strings.Contains(chapter.URL, qualityMap1[quality]) && !strings.Contains(chapter.URL, qualityMap2[quality]) {
				break
			} else {
				hasStream = true
			}
			duration, _ := strconv.ParseFloat(chapter.Duration, 64)
			stream.Segs = append(stream.Segs, proto.Seg{
				URL:      chapter.URL,
				Duration: duration,
			})
		}
		if hasStream {
			data.Streams = append(data.Streams, stream)
		}
	}
	return &data
}
