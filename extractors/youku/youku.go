package youku

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	netURL "net/url"
	"strings"
	"time"

	"github.com/iawia002/annie/extractors/types"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type errorData struct {
	Note string `json:"note"`
	Code int    `json:"code"`
}

type segs struct {
	Size int64  `json:"size"`
	URL  string `json:"cdn_url"`
}

type stream struct {
	Size      int64  `json:"size"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Segs      []segs `json:"segs"`
	Type      string `json:"stream_type"`
	AudioLang string `json:"audio_lang"`
}

type youkuVideo struct {
	Title string `json:"title"`
}

type youkuShow struct {
	Title string `json:"title"`
}

type data struct {
	Error  errorData  `json:"error"`
	Stream []stream   `json:"stream"`
	Video  youkuVideo `json:"video"`
	Show   youkuShow  `json:"show"`
}

type youkuData struct {
	Data data `json:"data"`
}

const youkuReferer = "https://v.youku.com"

func getAudioLang(lang string) string {
	var youkuAudioLang = map[string]string{
		"guoyu": "国语",
		"ja":    "日语",
		"yue":   "粤语",
	}
	translate, ok := youkuAudioLang[lang]
	if !ok {
		return lang
	}
	return translate
}

// https://g.alicdn.com/player/ykplayer/0.5.61/youku-player.min.js
// {"0505":"interior","050F":"interior","0501":"interior","0502":"interior","0503":"interior","0510":"adshow","0512":"BDskin","0590":"BDskin"}

// var ccodes = []string{"0510", "0502", "0507", "0508", "0512", "0513", "0514", "0503", "0590"}

func youkuUps(vid string, option types.Options) (*youkuData, error) {
	var (
		url   string
		utid  string
		utids []string
		data  youkuData
	)
	if strings.Contains(option.Cookie, "cna") {
		utids = utils.MatchOneOf(option.Cookie, `cna=(.+?);`, `cna\s+(.+?)\s`, `cna\s+(.+?)$`)

	} else {
		headers, err := request.Headers("http://log.mmstat.com/eg.js", youkuReferer)
		if err != nil {
			return nil, err
		}
		setCookie := headers.Get("Set-Cookie")
		utids = utils.MatchOneOf(setCookie, `cna=(.+?);`)
	}
	if utids == nil || len(utids) < 2 {
		return nil, types.ErrURLParseFailed
	}
	utid = utids[1]

	// https://g.alicdn.com/player/ykplayer/0.5.61/youku-player.min.js
	// grep -oE '"[0-9a-zA-Z+/=]{256}"' youku-player.min.js
	for _, ccode := range []string{option.YoukuCcode} {
		if ccode == "0103010102" {
			utid = generateUtdid()
		}
		url = fmt.Sprintf(
			"https://ups.youku.com/ups/get.json?vid=%s&ccode=%s&client_ip=192.168.1.1&client_ts=%d&utid=%s&ckey=%s",
			vid, ccode, time.Now().Unix()/1000, netURL.QueryEscape(utid), netURL.QueryEscape(option.YoukuCkey),
		)
		if option.YoukuPassword != "" {
			url = fmt.Sprintf("%s&password=%s", url, option.YoukuPassword)
		}
		html, err := request.GetByte(url, youkuReferer, nil)
		if err != nil {
			return nil, err
		}
		// data must be emptied before reassignment, otherwise it will contain the previous value(the 'error' data)
		data = youkuData{}
		if err = json.Unmarshal(html, &data); err != nil {
			return nil, err
		}
		if data.Data.Error == (errorData{}) {
			return &data, nil
		}
	}
	return &data, nil
}

func getBytes(val int32) []byte {
	var buff bytes.Buffer
	binary.Write(&buff, binary.BigEndian, val) // nolint
	return buff.Bytes()
}

func hashCode(s string) int32 {
	var result int32
	for _, c := range s {
		result = result*0x1f + c
	}
	return result
}

func hmacSha1(key []byte, msg []byte) []byte {
	mac := hmac.New(sha1.New, key)
	mac.Write(msg) // nolint
	return mac.Sum(nil)
}

func generateUtdid() string {
	timestamp := int32(time.Now().Unix())
	var buffer bytes.Buffer
	buffer.Write(getBytes(timestamp - 60*60*8))
	buffer.Write(getBytes(rand.Int31()))
	buffer.WriteByte(0x03)
	buffer.WriteByte(0x00)
	imei := fmt.Sprintf("%d", rand.Int31())
	buffer.Write(getBytes(hashCode(imei)))
	data := hmacSha1([]byte("d6fc3a4a06adbde89223bvefedc24fecde188aaa9161"), buffer.Bytes())
	buffer.Write(getBytes(hashCode(base64.StdEncoding.EncodeToString(data))))
	return base64.StdEncoding.EncodeToString(buffer.Bytes())
}

func genData(youkuData data) map[string]*types.Stream {
	var (
		streamString string
		quality      string
	)
	streams := make(map[string]*types.Stream, len(youkuData.Stream))
	for _, stream := range youkuData.Stream {
		if stream.AudioLang == "default" {
			streamString = stream.Type
			quality = fmt.Sprintf(
				"%s %dx%d", stream.Type, stream.Width, stream.Height,
			)
		} else {
			streamString = fmt.Sprintf("%s-%s", stream.Type, stream.AudioLang)
			quality = fmt.Sprintf(
				"%s %dx%d %s", stream.Type, stream.Width, stream.Height,
				getAudioLang(stream.AudioLang),
			)
		}

		ext := strings.Split(
			strings.Split(stream.Segs[0].URL, "?")[0],
			".",
		)
		urls := make([]*types.Part, len(stream.Segs))
		for index, data := range stream.Segs {
			urls[index] = &types.Part{
				URL:  data.URL,
				Size: data.Size,
				Ext:  ext[len(ext)-1],
			}
		}
		streams[streamString] = &types.Stream{
			Parts:   urls,
			Size:    stream.Size,
			Quality: quality,
		}
	}
	return streams
}

type extractor struct{}

// New returns a youtube extractor.
func New() types.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option types.Options) ([]*types.Data, error) {
	vids := utils.MatchOneOf(
		url, `id_(.+?)\.html`, `id_(.+)`,
	)
	if vids == nil || len(vids) < 2 {
		return nil, types.ErrURLParseFailed
	}
	vid := vids[1]

	youkuData, err := youkuUps(vid, option)
	if err != nil {
		return nil, err
	}
	if youkuData.Data.Error.Code != 0 {
		return nil, errors.New(youkuData.Data.Error.Note)
	}
	streams := genData(youkuData.Data)
	var title string
	if youkuData.Data.Show.Title == "" || strings.Contains(
		youkuData.Data.Video.Title, youkuData.Data.Show.Title,
	) {
		title = youkuData.Data.Video.Title
	} else {
		title = fmt.Sprintf("%s %s", youkuData.Data.Show.Title, youkuData.Data.Video.Title)
	}

	return []*types.Data{
		{
			Site:    "优酷 youku.com",
			Title:   title,
			Type:    types.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
