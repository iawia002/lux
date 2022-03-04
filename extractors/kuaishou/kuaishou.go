package kuaishou

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("kuaishou", New())
}

type extractor struct{}

// New returns a kuaishou extractor.
func New() extractors.Extractor {
	return &extractor{}
}

// fetch url and get the cookie that write by server
func fetchCookies(url string, headers map[string]string) (string, error) {
	res, err := request.Request(http.MethodGet, url, nil, headers)
	if err != nil {
		return "", err
	}

	defer res.Body.Close() // nolint

	cookiesArr := make([]string, 0)
	cookies := res.Cookies()

	for _, c := range cookies {
		cookiesArr = append(cookiesArr, c.Name+"="+c.Value)
	}

	return strings.Join(cookiesArr, "; "), nil
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:98.0) Gecko/20100101 Firefox/98.0",
	}

	cookies, err := fetchCookies(url, headers)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	headers["Cookie"] = cookies

	html, err := request.Get(url, url, headers)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	titles := utils.MatchOneOf(html, `<title>([^<]+)</title>`)
	if titles == nil || len(titles) < 2 {
		return nil, errors.New("can not found title")
	}

	title := regexp.MustCompile(`\n+`).ReplaceAllString(strings.TrimSpace(titles[1]), " ")

	qualityRegMap := map[string]*regexp.Regexp{
		"sd": regexp.MustCompile(`"photoUrl":\s*"([^"]+)"`),
	}

	streams := make(map[string]*extractors.Stream, 1)
	for quality, qualityReg := range qualityRegMap {
		matcher := qualityReg.FindStringSubmatch(html)
		if len(matcher) != 2 {
			return nil, errors.WithStack(extractors.ErrURLParseFailed)
		}

		u := strings.ReplaceAll(matcher[1], `\u002F`, "/")

		size, err := request.Size(u, url)
		if err != nil {
			return nil, errors.WithStack(err)
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
			Site:    "快手 kuaishou.com",
			Title:   title,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
