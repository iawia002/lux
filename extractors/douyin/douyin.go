package douyin

import (
	"crypto/rand"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	netURL "net/url"
	"regexp"
	"strings"

	"github.com/dop251/goja"
	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	e := New()
	extractors.Register("douyin", e)
	extractors.Register("iesdouyin", e)
}

//go:embed sign.js
var script string

type extractor struct{}

// New returns a douyin extractor.
func New() extractors.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	if strings.Contains(url, "v.douyin.com") {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		c := http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		resp, err := c.Do(req)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		defer resp.Body.Close() // nolint
		url = resp.Header.Get("location")
	}

	itemIds := utils.MatchOneOf(url, `/video/(\d+)`)
	if len(itemIds) == 0 {
		return nil, errors.New("unable to get video ID")
	}
	if itemIds == nil || len(itemIds) < 2 {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}
	itemId := itemIds[len(itemIds)-1]

	// dynamic generate cookie
	cookie, err := createCookie()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	api := "https://www.douyin.com/aweme/v1/web/aweme/detail/?aweme_id=" + itemId
	// parse api query params string
	query, err := netURL.Parse(api)
	if err != nil {
		return nil, errors.WithStack(extractors.ErrURLQueryParamsParseFailed)
	}
	// define request headers and sign agent
	headers := map[string]string{}
	headers["Cookie"] = cookie
	headers["Referer"] = "https://www.douyin.com/"
	headers["User-Agent"] = "Mozilla/5.0 (Linux; Android 8.0; Pixel 2 Build/OPD3.170816.012) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Mobile Safari/537.36 Edg/87.0.664.66"

	// init JavaScripts runtime
	vm := goja.New()
	// load sign scripts
	_, _ = vm.RunString(script)
	// sign
	sign, err := vm.RunString(fmt.Sprintf("sign('%s', '%s')", query.RawQuery, headers["User-Agent"]))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	api = fmt.Sprintf("%s&X-Bogus=%s", api, sign)

	jsonData, err := request.Get(api, url, headers)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var douyin douyinData
	if err = json.Unmarshal([]byte(jsonData), &douyin); err != nil {
		return nil, errors.WithStack(err)
	}

	urlData := make([]*extractors.Part, 0)
	var douyinType extractors.DataType
	var totalSize int64
	// AwemeType: 0:video 68:image
	if douyin.AwemeDetail.AwemeType == 68 {
		douyinType = extractors.DataTypeImage
		for _, img := range douyin.AwemeDetail.Images {
			realURL := img.URLList[len(img.URLList)-1]
			size, err := request.Size(realURL, url)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			totalSize += size
			_, ext, err := utils.GetNameAndExt(realURL)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			urlData = append(urlData, &extractors.Part{
				URL:  realURL,
				Size: size,
				Ext:  ext,
			})
		}
	} else {
		douyinType = extractors.DataTypeVideo
		realURL := douyin.AwemeDetail.Video.PlayAddr.URLList[0]
		totalSize, err = request.Size(realURL, url)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		urlData = append(urlData, &extractors.Part{
			URL:  realURL,
			Size: totalSize,
			Ext:  "mp4",
		})
	}
	streams := map[string]*extractors.Stream{
		"default": {
			Parts: urlData,
			Size:  totalSize,
		},
	}

	return []*extractors.Data{
		{
			Site:    "抖音 douyin.com",
			Title:   douyin.AwemeDetail.Desc,
			Type:    douyinType,
			Streams: streams,
			URL:     url,
		},
	}, nil
}

func createCookie() (string, error) {
	v1, err := msToken(107)
	if err != nil {
		return "", err
	}
	v2, err := ttwid()
	if err != nil {
		return "", err
	}
	v3 := "324fb4ea4a89c0c05827e18a1ed9cf9bf8a17f7705fcc793fec935b637867e2a5a9b8168c885554d029919117a18ba69"
	v4 := "eyJiZC10aWNrZXQtZ3VhcmQtdmVyc2lvbiI6MiwiYmQtdGlja2V0LWd1YXJkLWNsaWVudC1jc3IiOiItLS0tLUJFR0lOIENFUlRJRklDQVRFIFJFUVVFU1QtLS0tLVxyXG5NSUlCRFRDQnRRSUJBREFuTVFzd0NRWURWUVFHRXdKRFRqRVlNQllHQTFVRUF3d1BZbVJmZEdsamEyVjBYMmQxXHJcbllYSmtNRmt3RXdZSEtvWkl6ajBDQVFZSUtvWkl6ajBEQVFjRFFnQUVKUDZzbjNLRlFBNUROSEcyK2F4bXAwNG5cclxud1hBSTZDU1IyZW1sVUE5QTZ4aGQzbVlPUlI4NVRLZ2tXd1FJSmp3Nyszdnc0Z2NNRG5iOTRoS3MvSjFJc3FBc1xyXG5NQ29HQ1NxR1NJYjNEUUVKRGpFZE1Cc3dHUVlEVlIwUkJCSXdFSUlPZDNkM0xtUnZkWGxwYmk1amIyMHdDZ1lJXHJcbktvWkl6ajBFQXdJRFJ3QXdSQUlnVmJkWTI0c0RYS0c0S2h3WlBmOHpxVDRBU0ROamNUb2FFRi9MQnd2QS8xSUNcclxuSURiVmZCUk1PQVB5cWJkcytld1QwSDZqdDg1czZZTVNVZEo5Z2dmOWlmeTBcclxuLS0tLS1FTkQgQ0VSVElGSUNBVEUgUkVRVUVTVC0tLS0tXHJcbiJ9"
	cookie := fmt.Sprintf("msToken=%s;ttwid=%s;odin_tt=%s;bd_ticket_guard_client_data=%s;", v1, v2, v3, v4)
	return cookie, nil
}

func msToken(length int) (string, error) {
	const characters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	randomBytes := make([]byte, length)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	token := make([]byte, length)
	for i, b := range randomBytes {
		token[i] = characters[int(b)%len(characters)]
	}
	return string(token), nil
}

func ttwid() (string, error) {
	body := map[string]interface{}{
		"aid":           1768,
		"union":         true,
		"needFid":       false,
		"region":        "cn",
		"cbUrlProtocol": "https",
		"service":       "www.ixigua.com",
		"migrate_info":  map[string]string{"ticket": "", "source": "node"},
	}
	bytes, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	payload := strings.NewReader(string(bytes))
	resp, err := request.Request(http.MethodPost, "https://ttwid.bytedance.com/ttwid/union/register/", payload, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close() // nolint
	cookie := resp.Header.Get("Set-Cookie")
	re := regexp.MustCompile(`ttwid=([^;]+)`)
	if match := re.FindStringSubmatch(cookie); match != nil {
		return match[1], nil
	}
	return "", errors.New("douyin ttwid request failed")
}
