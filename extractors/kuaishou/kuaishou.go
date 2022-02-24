package kuaishou

import (
	"compress/flate"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	stdUrl "net/url"
	"regexp"
	"strings"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("kuaishou", New())
}

type extractor struct{}

// New returns a facebook extractor.
func New() extractors.Extractor {
	return &extractor{}
}

func requestHTML(url string, retryCount int, headers map[string]string) (htmlStr string, cookie string, err error) {
	if headers == nil {
		headers = make(map[string]string)
	}

	if retryCount > 1 {
		return "", "", errors.New("retry timeout")
	}

	u := url
	if retryCount == 0 {
		u = "https://www.kuaishou.com"
	}

	res, err := request.Request(http.MethodGet, u, nil, headers)
	if err != nil {
		return "", "", err
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

	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", "", err
	}

	htmlStr = string(b)

	titles := utils.MatchOneOf(htmlStr, `<title>([^<]+)</title>`)
	if titles == nil || len(titles) < 2 {
		return "", "", extractors.ErrURLParseFailed
	}

	title := strings.TrimSpace(titles[1])

	cookiesArr := make([]string, 0)
	cookies := res.Cookies()

	for _, c := range cookies {
		cookiesArr = append(cookiesArr, c.Name+"="+c.Value)
	}

	cookieStr := strings.Join(cookiesArr, "; ")

	if title == "短视频-快手" || strings.Contains(title, "快手短视频") {
		retryCount += 1
		// headers["Cookie"] = "kpf=PC_WEB; kpn=KUAISHOU_VISION; clientid=3; did=web_3931f3e6038ba69679184dcbb2e9c2dc"
		headers["Cookie"] = cookieStr
		return requestHTML(url, retryCount, headers)
	}

	return htmlStr, cookieStr, nil
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	var err error

	inputUrl, err := stdUrl.Parse(url)
	if err != nil {
		return nil, err
	}

	html, _, err := requestHTML(url, 0, map[string]string{
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:98.0) Gecko/20100101 Firefox/98.0",
		"Origin":     inputUrl.Scheme + "://" + inputUrl.Host, // https://github.com
		"Host":       inputUrl.Hostname(),                     // github.com
	})
	if err != nil {
		return nil, err
	}

	titles := utils.MatchOneOf(html, `<title>([^<]+)</title>`)
	if titles == nil || len(titles) < 2 {
		return nil, errors.New("can not found title")
	}

	title := strings.TrimSpace(titles[1])

	title = regexp.MustCompile(`\n+`).ReplaceAllString(title, " ")

	qualityRegMap := map[string]*regexp.Regexp{
		"sd": regexp.MustCompile(`"photoUrl":\s*"([^"]+)"`),
	}

	streams := make(map[string]*extractors.Stream, 2)
	for quality, qualityReg := range qualityRegMap {
		matcher := qualityReg.FindStringSubmatch(html)

		if len(matcher) != 2 {
			return nil, extractors.ErrURLParseFailed
		}

		u := strings.ReplaceAll(matcher[1], `\u002F`, "/")

		size, err := request.Size(u, url)
		if err != nil {
			return nil, err
		}

		urlData := &extractors.Part{
			URL:  u,
			Size: size,
			Ext:  "mp4",
		}
		streams[quality] = &extractors.Stream{
			Parts:   []*extractors.Part{urlData},
			Size:    size,
			Quality: quality,
		}
	}

	return []*extractors.Data{
		{
			Site:    "KuaiShou www.kuaishou.com",
			Title:   title,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
