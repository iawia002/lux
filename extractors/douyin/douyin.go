package douyin

import (
	"encoding/json"
	"errors"
	"github.com/iawia002/annie/extractors/types"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
	"strings"
)

type extractor struct{}

// New returns a youtube extractor.
func New() types.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option types.Options) ([]*types.Data, error) {
	var err error
	itemIds := utils.MatchOneOf(url, `/video/(\d+)`)
	if len(itemIds) == 0 {
		return nil, errors.New("unable to get video ID")
	}

	if itemIds == nil || len(itemIds) < 2 {
		return nil, types.ErrURLParseFailed
	}
	itemId := itemIds[len(itemIds)-1]
	jsonData, err := request.Get("https://www.iesdouyin.com/web/api/v2/aweme/iteminfo/?item_ids="+itemId, url, nil)
	var douyin douyinData
	err = json.Unmarshal([]byte(jsonData), &douyin)
	if err != nil {
		return nil, err
	}
	realURL := strings.Replace(douyin.ItemList[0].Video.PlayAddr.URLList[0], "playwm", "play", -1)
	size, err := request.Size(realURL, url)
	if err != nil {
		return nil, err
	}
	urlData := &types.Part{
		URL:  realURL,
		Size: size,
		Ext:  "mp4",
	}
	streams := map[string]*types.Stream{
		"default": {
			Parts: []*types.Part{urlData},
			Size:  size,
		},
	}

	return []*types.Data{
		{
			Site:    "抖音 douyin.com",
			Title:   douyin.ItemList[0].Desc,
			Type:    types.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
