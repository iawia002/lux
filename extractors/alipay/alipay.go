package alipay

import (
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
)

const webGW = "https://webgw-internet.alipay.com/contentservice/com.alipay.sofa.function.SOFAFunction/apply/myjf.yuyan.contentservice.default.ContentWebGWController.list"

var clarities = []string{"1080P_h265_EH", "1080P_h265", "720P_h265", "720P_h264", "540P_h264"}
var idPattern = regexp.MustCompile(`(?:[?&]|\b)contentId=([A-Za-z0-9]+)(?:[^A-Za-z0-9]|$)`)

type extractor struct {
	request func(string, string, io.Reader, map[string]string) (*http.Response, error)
}

func init() {
	e := New()
	extractors.Register("lepudding", e)
	extractors.Register("alipay", e)
}

// New returns an Alipay public video extractor.
func New() extractors.Extractor { return &extractor{request: request.Request} }

func contentID(text string) string {
	for i := 0; i < 4; i++ {
		if match := idPattern.FindStringSubmatch(text); len(match) > 1 {
			return match[1]
		}
		decoded, err := url.PathUnescape(text)
		if err != nil || decoded == text {
			break
		}
		text = decoded
	}
	return ""
}

func (e *extractor) resolveID(link string) (string, error) {
	if id := contentID(link); id != "" {
		return id, nil
	}
	resp, err := e.request(http.MethodGet, link, nil, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.Request != nil {
		if id := contentID(resp.Request.URL.String()); id != "" {
			return id, nil
		}
	}
	body, err := responseBody(resp)
	if err != nil {
		return "", err
	}
	if id := contentID(string(body)); id != "" {
		return id, nil
	}
	return "", fmt.Errorf("alipay: could not find contentId in share URL")
}

func responseBody(resp *http.Response) ([]byte, error) {
	var reader io.Reader = resp.Body
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		decoded, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer decoded.Close()
		reader = decoded
	case "deflate":
		decoded := flate.NewReader(resp.Body)
		defer decoded.Close()
		reader = decoded
	}
	return io.ReadAll(io.LimitReader(reader, 4<<20))
}

func streamURLs(original string) (map[string]string, error) {
	u, err := url.Parse(original)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, fmt.Errorf("alipay: invalid video URL")
	}
	streams := map[string]string{"original": original}
	if !strings.Contains(u.Path, "/") {
		return streams, nil
	}
	index := strings.LastIndex(u.Path, "/")
	for _, clarity := range clarities {
		if u.Path[index+1:] == clarity {
			delete(streams, "original")
			for _, candidate := range clarities {
				v := *u
				v.Path = u.Path[:index+1] + candidate
				v.RawPath = ""
				streams[candidate] = v.String()
			}
			break
		}
	}
	return streams, nil
}

func (e *extractor) streamSize(link string) (int64, error) {
	resp, err := e.request(http.MethodGet, link, nil, map[string]string{"Range": "bytes=0-0"})
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return 0, fmt.Errorf("alipay: stream returned HTTP %d", resp.StatusCode)
	}
	length := resp.Header.Get("Content-Length")
	if resp.StatusCode == http.StatusPartialContent {
		_, length, _ = strings.Cut(resp.Header.Get("Content-Range"), "/")
	}
	size, err := strconv.ParseInt(length, 10, 64)
	if err != nil || size <= 0 {
		return 0, fmt.Errorf("alipay: invalid stream size %q", length)
	}
	return size, nil
}

// Extract resolves a public share link and discovers available video qualities.
func (e *extractor) Extract(link string, option extractors.Options) ([]*extractors.Data, error) {
	id, err := e.resolveID(link)
	if err != nil {
		return nil, err
	}
	payload, _ := json.Marshal(map[string]string{"contentId": id})
	resp, err := e.request(http.MethodPost, webGW, strings.NewReader(string(payload)), map[string]string{
		"x-webgw-appId":   "180020010001266490",
		"x-webgw-version": "2.0",
		"Content-Type":    "application/json",
		"User-Agent":      "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15",
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var detail struct {
		Success      bool   `json:"success"`
		ErrorMessage string `json:"errorMessage"`
		ResultObj    struct {
			Data []struct {
				Title string `json:"title"`
				Video struct {
					VID string `json:"vid"`
				} `json:"video"`
			} `json:"data"`
		} `json:"resultObj"`
	}
	body, err := responseBody(resp)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &detail); err != nil {
		return nil, fmt.Errorf("alipay: invalid API response: %w", err)
	}
	if !detail.Success {
		return nil, fmt.Errorf("alipay: API error: %s", detail.ErrorMessage)
	}
	if len(detail.ResultObj.Data) == 0 || detail.ResultObj.Data[0].Video.VID == "" {
		return nil, fmt.Errorf("alipay: no video in API response")
	}
	item := detail.ResultObj.Data[0]
	candidates, err := streamURLs(item.Video.VID)
	if err != nil {
		return nil, err
	}
	streams := make(map[string]*extractors.Stream)
	for quality, streamURL := range candidates {
		size, err := e.streamSize(streamURL)
		if err != nil {
			continue
		}
		streams[quality] = &extractors.Stream{
			Quality: quality, Size: size,
			Parts: []*extractors.Part{{URL: streamURL, Size: size, Ext: "mp4"}},
		}
	}
	if len(streams) == 0 {
		return nil, fmt.Errorf("alipay: no available video streams")
	}
	title := item.Title
	if title == "" {
		title = id
	}
	return []*extractors.Data{{URL: link, Site: "支付宝 Alipay", Title: title, Type: extractors.DataTypeVideo, Streams: streams}}, nil
}
