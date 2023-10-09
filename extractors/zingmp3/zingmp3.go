package zingmp3

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"maps"
	"net/http"
	neturl "net/url"
	"regexp"
	"sort"

	"github.com/buger/jsonparser"
	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	zingmp3Extractor := New()
	extractors.Register("zingmp3", zingmp3Extractor)
	extractors.Register("zing", zingmp3Extractor)
}

type extractor struct{}

// New returns a zingmp3 extractor.
func New() extractors.Extractor {
	return &extractor{}
}

type params map[string]string

var ApiSlugs = map[string]string{
	"bai-hat":        "/api/v2/page/get/song",
	"embed":          "/api/v2/page/get/song",
	"video-clip":     "/api/v2/page/get/video",
	"lyric":          "/api/v2/lyric/get/lyric",
	"song-streaming": "/api/v2/song/get/streaming",
}

const Domain = "https://zingmp3.vn"

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	urlRegExp := regexp.MustCompile(`https?://(?:mp3\.zing|zingmp3)\.vn/(?P<type>(?:bai-hat|video-clip|embed))/[^/?#]+/(?P<id>\w+)(?:\.html|\?)`)
	urlMatcher := urlRegExp.FindStringSubmatch(url)
	if len(urlMatcher) == 0 {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}
	urlType := urlMatcher[1]
	id := urlMatcher[2]
	if err := updatingCookies(); err != nil {
		return nil, errors.WithStack(err)
	}
	data := callApi(urlType, params{"id": id})
	title, _ := jsonparser.GetString(data, "title")
	var contentType extractors.DataType
	var source []byte
	if urlType == "video-clip" {
		source, _, _, _ = jsonparser.Get(data, "streaming")
		api := fmt.Sprintf(`http://api.mp3.zing.vn/api/mobile/video/getvideoinfo?requestdata={"id":"%s"}`, id)
		res, _ := request.Get(api, api, nil)
		newSource, _, _, _ := jsonparser.Get([]byte(res), "source")
		source, _ = jsonparser.Set(source, newSource, "mp4")
		contentType = extractors.DataTypeVideo
	} else {
		contentType = extractors.DataTypeAudio
		source = callApi("song-streaming", params{"id": id})
	}
	streams := make(map[string]*extractors.Stream)
	if err := jsonparser.ObjectEach(source, func(k []byte, v []byte, dataType jsonparser.ValueType, offset int) error {
		key := string(k)
		value := string(v)
		if value == "" || value == "VIP" {
			return nil
		}

		// Handle for audio
		if key != "mp4" && key != "hls" {
			size, _ := request.Size(value, url)
			urlData := &extractors.Part{
				URL:  value,
				Ext:  "mp3",
				Size: size,
			}
			streams["default"] = &extractors.Stream{
				Parts: []*extractors.Part{urlData},
			}
			return nil
		}

		// Handle for video
		return jsonparser.ObjectEach(v, func(kSource []byte, vSource []byte, _ jsonparser.ValueType, _ int) error {
			resolution := string(kSource)
			videoUrl := string(vSource)
			if resolution == "" {
				return nil
			}
			if resolution == "hls" {
				urls, _ := utils.M3u8URLs(videoUrl)
				parts := make([]*extractors.Part, 0)
				for _, u := range urls {
					parts = append(parts, &extractors.Part{
						URL: u,
						Ext: "ts",
					})
				}
				streams[resolution] = &extractors.Stream{
					ID:      resolution,
					Parts:   parts,
					NeedMux: false,
				}
				return nil
			}
			size, _ := request.Size(videoUrl, url)
			streams[fmt.Sprintf("mp4-%s", resolution)] = &extractors.Stream{
				Parts: []*extractors.Part{{
					URL:  videoUrl,
					Ext:  "mp4",
					Size: size,
				}},
			}
			return nil
		})
	}); err != nil {
		return nil, errors.WithStack(err)
	}

	return []*extractors.Data{
		{
			Site:    "Zing MP3 zingmp3.vn",
			Title:   title,
			Type:    contentType,
			Streams: streams,
			URL:     url,
		},
	}, nil
}

func callApi(urlType string, p params) []byte {
	api := generateApi(urlType, p)
	res, _ := request.GetByte(api, api, nil)
	data, _, _, _ := jsonparser.Get(res, "data")
	return data
}

func updatingCookies() error {
	// For the first time. We need to call the temp API to get cookies and set cookies to for next request
	// But sometime zingmp3 doesn't return cookies. We need to retry get and set cookies again (only allow 5 time)
	for i := 0; i < 5; i++ {
		api := generateApi("bai-hat", params{"id": ""})
		res, err := request.Request(http.MethodGet, api, nil, nil)
		if err != nil {
			return err
		}
		cookies := ""
		for _, value := range res.Cookies() {
			cookies += value.String()
		}
		res.Body.Close() // nolint
		if cookies != "" {
			request.SetOptions(request.Options{
				Cookie: cookies,
			})
			return nil
		}
	}
	return nil
}

func generateApi(urlType string, p params) string {
	slugApi := ApiSlugs[urlType]
	maps.Copy(p, params{"ctime": "1"})

	sortedParams := sortedParams(p)
	sig := generateSig(slugApi, sortedParams)
	maps.Copy(sortedParams, params{
		"apiKey": "X5BM3w8N7MKozC0B85o4KMlzLZKhV00y",
		"sig":    sig,
	})

	urlParams := neturl.Values{}
	for key, value := range sortedParams {
		urlParams.Add(key, value)
	}
	return fmt.Sprintf("%s%s?%s", Domain, slugApi, urlParams.Encode())
}

func generateSig(slugApi string, p params) string {
	str := ""
	for key, value := range p {
		str += fmt.Sprintf("%s=%s", key, value)
	}
	h := sha256.New()
	h.Write([]byte(str))
	sha256Value := hex.EncodeToString(h.Sum(nil))
	var passwordBytes = []byte(fmt.Sprintf("%s%s", slugApi, sha256Value))
	salt := []byte("acOrvUS15XRW2o9JksiK1KgQ6Vbds8ZW")
	hmacHashed := hmac.New(sha512.New, salt)
	hmacHashed.Write(passwordBytes)
	return hex.EncodeToString(hmacHashed.Sum(nil))
}

func sortedParams(p params) params {
	keys := make([]string, 0, len(p))
	for k := range p {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	sortedParams := params{}
	for _, k := range keys {
		sortedParams[k] = p[k]
	}
	return sortedParams
}
