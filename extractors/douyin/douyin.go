package douyin

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/iawia002/lux/extractors/types"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

type extractor struct{}

// New returns a douyin extractor.
func New() types.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option types.Options) ([]*types.Data, error) {
	if strings.Contains(url, "v.douyin.com") {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		c := http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		resp, err := c.Do(req)
		if err != nil {
			return nil, err
		}
		url = resp.Header.Get("location")
	}

	itemIds := utils.MatchOneOf(url, `/video/(\d+)`)
	if len(itemIds) == 0 {
		return nil, errors.New("unable to get video ID")
	}
	if itemIds == nil || len(itemIds) < 2 {
		return nil, types.ErrURLParseFailed
	}
	itemId := itemIds[len(itemIds)-1]
	jsonData, err := request.Get("https://www.iesdouyin.com/web/api/v2/aweme/iteminfo/?item_ids="+itemId, url, nil)
	if err != nil {
		return nil, err
	}
	var douyin douyinData
	if err = json.Unmarshal([]byte(jsonData), &douyin); err != nil {
		return nil, err
	}

	urlData := make([]*types.Part, 0)
	var douyinType types.DataType
	var totalSize int64
	// AwemeType: 2:image 4:video
	if douyin.ItemList[0].AwemeType == 2 {
		douyinType = types.DataTypeImage
		for _, img := range douyin.ItemList[0].Images {
			realURL := img.URLList[len(img.URLList)-1]
			size, err := request.Size(realURL, url)
			if err != nil {
				return nil, err
			}
			totalSize += size
			_, ext, err := utils.GetNameAndExt(realURL)
			if err != nil {
				return nil, err
			}
			urlData = append(urlData, &types.Part{
				URL:  realURL,
				Size: size,
				Ext:  ext,
			})
		}
	} else {
		douyinType = types.DataTypeVideo
		realURL := "https://aweme.snssdk.com/aweme/v1/play/?video_id=" + douyin.ItemList[0].Video.PlayAddr.URI + "&ratio=720p&line=0"
		totalSize, err = request.Size(realURL, url)
		if err != nil {
			return nil, err
		}
		urlData = append(urlData, &types.Part{
			URL:  realURL,
			Size: totalSize,
			Ext:  "mp4",
		})
	}
	streams := map[string]*types.Stream{
		"default": {
			Parts: urlData,
			Size:  totalSize,
		},
	}

	return []*types.Data{
		{
			Site:    "抖音 douyin.com",
			Title:   douyin.ItemList[0].Desc,
			Type:    douyinType,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
