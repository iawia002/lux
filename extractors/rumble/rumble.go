package rumble

import (
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"

	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("rumble", New())
}

type extractor struct{}

// New returns a rumble extractor.
func New() extractors.Extractor {
	return &extractor{}
}

type rumbleData struct {
	Format       string `json:"format"`
	Name         string `json:"name"`
	EmbedURL     string `json:"embedUrl"`
	ThumbnailURL string `json:"thumbnailUrl"`
	Type         string `json:"@type"`
	VideoURL     string `json:"videoUrl"`
	Quality      string `json:"quality"`
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	res, err := request.Request(http.MethodGet, url, nil, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	defer res.Body.Close() // nolint

	var reader io.ReadCloser
	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		reader, _ = gzip.NewReader(res.Body)
	case "deflate":
		reader = flate.NewReader(res.Body)
	default:
		reader = res.Body
	}
	defer reader.Close() // nolint

	b, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	html := string(b)
	var title string
	matchTitle := utils.MatchOneOf(html, `<title>(.+?)</title>`)
	if len(matchTitle) > 1 {
		title = matchTitle[1]
	} else {
		title = "rumble video"
	}

	payload, err := readPayload(html)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	videoID, err := getVideoID(payload.EmbedURL)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	streams, err := fetchVideoQuality(videoID)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return []*extractors.Data{
		{
			Site:    "Rumble rumble.com",
			Title:   title,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}

// Read JSON object from the video webpage
func readPayload(html string) (*rumbleData, error) {
	matchPayload := utils.MatchOneOf(html, `\<script\stype="?application\/ld\+json"?\>(.+?)\<\/script>`)
	if len(matchPayload) < 1 {
		return nil, errors.WithStack(extractors.ErrURLQueryParamsParseFailed)
	}

	rumbles := make([]rumbleData, 0)
	if err := json.Unmarshal([]byte(matchPayload[1]), &rumbles); err != nil {
		return nil, errors.WithStack(err)
	}

	for _, it := range rumbles {
		if it.Type == "VideoObject" {
			return &it, nil
		}
	}

	return nil, errors.WithStack(extractors.ErrURLParseFailed)
}

func getVideoID(embedURL string) (string, error) {
	u, err := url.Parse(embedURL)
	if err != nil {
		return "", errors.WithStack(extractors.ErrURLParseFailed)
	}

	return path.Base(u.Path), nil
}

// Rumble response contains the streams in `rumbleStreams`
type rumbleResponse struct {
	Streams *json.RawMessage `json:"ua"`
}

// Common video meta data
type streamInfo struct {
	URL  string `json:"url"`
	Meta struct {
		Bitrate uint16 `json:"bitrate"`
		Size    int64  `json:"size"`
		Width   uint16 `json:"w"`
		Height  uint16 `json:"h"`
	} `json:"meta"`
}

// common video qualities for `mp4`, `webm`
type videoQualities struct {
	Q240  struct{ streamInfo } `json:"240"`
	Q360  struct{ streamInfo } `json:"360"`
	Q480  struct{ streamInfo } `json:"480"`
	Q720  struct{ streamInfo } `json:"720"`
	Q1080 struct{ streamInfo } `json:"1080"`
	Q1440 struct{ streamInfo } `json:"1440"`
	Q2160 struct{ streamInfo } `json:"2160"`
	Q2161 struct{ streamInfo } `json:"2161"`
}

// Video payload for adaptive stream and different qualities
type rumbleStreams struct {
	FMp4 struct {
		videoQualities
	} `json:"mp4"`
	FWebm struct {
		videoQualities
	} `json:"webm"`
	FHLS struct {
		QAuto struct{ streamInfo } `json:"auto"`
	} `json:"hls"`
}

// Unmarshall the video response
// Some properties like `mp4`, `webm` are either array or an object
func (r *rumbleStreams) UnmarshalJSON(b []byte) error {
	var resp *rumbleResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		return errors.WithStack(extractors.ErrURLParseFailed)
	}

	// Get individual stream from the response
	var obj map[string]*json.RawMessage
	if err := json.Unmarshal(*resp.Streams, &obj); err != nil {
		return errors.WithStack(extractors.ErrURLParseFailed)
	}

	if v, ok := obj["mp4"]; ok {
		_ = json.Unmarshal(*v, &r.FMp4)
	}
	if v, ok := obj["webm"]; ok {
		_ = json.Unmarshal(*v, &r.FWebm)
	}
	if v, ok := obj["hls"]; ok {
		_ = json.Unmarshal(*v, &r.FHLS)
	}

	return nil
}

// Use this to create all the streams for `mp4`, `webm`
func (rs *rumbleStreams) makeAllVODStreams(m map[string]*extractors.Stream) {
	m["webm"] = makeStreamMeta("480", "webm", &rs.FWebm.Q480.streamInfo)
	m["240"] = makeStreamMeta("240", "mp4", &rs.FMp4.Q240.streamInfo)
	m["360"] = makeStreamMeta("360", "mp4", &rs.FMp4.Q360.streamInfo)
	m["480"] = makeStreamMeta("480", "mp4", &rs.FMp4.Q480.streamInfo)
	m["720"] = makeStreamMeta("720", "mp4", &rs.FMp4.Q720.streamInfo)
	m["1080"] = makeStreamMeta("1080", "mp4", &rs.FMp4.Q1080.streamInfo)
	m["1440"] = makeStreamMeta("1440", "mp4", &rs.FMp4.Q1440.streamInfo)
	m["2160"] = makeStreamMeta("2160", "mp4", &rs.FMp4.Q2160.streamInfo)
	m["2161"] = makeStreamMeta("2161", "mp4", &rs.FMp4.Q2161.streamInfo)
}

var reResolution = regexp.MustCompile(`_(\d{3,4})p\/`) // ex. _720p/

// Use this to create all the streams for live videos
func (rs *rumbleStreams) makeAllLiveStreams(m map[string]*extractors.Stream) error {
	playlists, err := utils.M3u8URLs(rs.FHLS.QAuto.URL)
	if err != nil {
		return errors.WithStack(err)
	}

	if len(playlists) == 0 {
		return errors.WithStack(extractors.ErrURLParseFailed)
	}

	// Find the highest resolution
	playlistURL := playlists[0]
	maxRes := 0
	for _, x := range playlists {
		matched := reResolution.FindStringSubmatch(x)
		if len(matched) == 0 {
			continue
		}
		res, err := strconv.Atoi(matched[1])
		if err != nil {
			continue
		}

		if maxRes < res {
			maxRes = res
			playlistURL = x
		}
	}

	tsURLs, err := utils.M3u8URLs(playlistURL)
	if err != nil {
		return errors.WithStack(err)
	}

	var parts []*extractors.Part
	for _, x := range tsURLs {
		part := &extractors.Part{
			URL:  x,
			Size: rs.FHLS.QAuto.streamInfo.Meta.Size,
			Ext:  "ts",
		}
		parts = append(parts, part)
	}

	m["hls"] = &extractors.Stream{
		Parts:   parts,
		Size:    rs.FHLS.QAuto.streamInfo.Meta.Size,
		Quality: strconv.Itoa(maxRes),
	}

	return nil
}

// Request video formats and qualities
func fetchVideoQuality(videoID string) (map[string]*extractors.Stream, error) {
	reqURL := fmt.Sprintf(`https://rumble.com/embedJS/u3/?request=video&ver=2&v=%s&ext={"ad_count":null}&ad_wt=0`, videoID)

	res, err := request.Request(http.MethodGet, reqURL, nil, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer res.Body.Close() // nolint

	var reader io.ReadCloser
	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		reader, _ = gzip.NewReader(res.Body)
	case "deflate":
		reader = flate.NewReader(res.Body)
	default:
		reader = res.Body
	}
	defer reader.Close() // nolint

	b, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var rs rumbleStreams
	if err := json.Unmarshal(b, &rs); err != nil {
		return nil, errors.WithStack(err)
	}

	streams := make(map[string]*extractors.Stream, 9)
	rs.makeAllVODStreams(streams)
	_ = rs.makeAllLiveStreams(streams)

	return streams, nil
}

func makeStreamMeta(q, ext string, info *streamInfo) *extractors.Stream {
	urlMeta := &extractors.Part{
		URL:  info.URL,
		Size: info.Meta.Size,
		Ext:  ext,
	}

	return &extractors.Stream{
		Parts:   []*extractors.Part{urlMeta},
		Size:    info.Meta.Size,
		Quality: q,
	}
}
