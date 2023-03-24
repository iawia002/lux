package douyin

import (
	"encoding/json"
	"fmt"
	"github.com/dop251/goja"
	"net/http"
	netURL "net/url"
	"os"
	"strings"

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

	// init JavaScripts runtime
	vm := goja.New()
	// read sign file
	script, err := os.ReadFile("./script/douyin-sign.js")
	if err != nil {
		return nil, errors.WithStack(extractors.ErrSignScriptsRead)
	}
	api := "https://www.douyin.com/aweme/v1/web/aweme/detail/?aweme_id=" + itemId
	// parse api query params string
	query, err := netURL.Parse(api)
	if err != nil {
		return nil, errors.WithStack(extractors.ErrURLQueryParamsParseFailed)
	}
	// define request headers and sign agent
	headers := map[string]string{}
	headers["Cookie"] = "s_v_web_id=verify_leytkxgn_kvO5kOmO_SdMs_4t1o_B5ml_BUqtWM1mP6BF;"
	agent := "Mozilla/5.0 (Linux; Android 8.0; Pixel 2 Build/OPD3.170816.012) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Mobile Safari/537.36 Edg/87.0.664.66"
	// load sign scripts
	_, _ = vm.RunString(string(script))
	// sign
	sign, err := vm.RunString(fmt.Sprintf("sign('%s', '%s')", query.RawQuery, agent))
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
