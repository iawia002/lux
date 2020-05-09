package douyin

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iawia002/annie/extractors/types"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type data struct {
	ItemList []struct {
		Desc string `json:"desc"`
	} `json:"item_list"`
}

type extractor struct{}

// New returns a youtube extractor.
func New() types.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option types.Options) ([]*types.Data, error) {
	var err error
	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, err
	}

	realURLs := utils.MatchOneOf(html, `playAddr: "(.+?)"`)
	if realURLs == nil || len(realURLs) < 2 {
		return nil, types.ErrURLParseFailed
	}
	realURL := realURLs[1]

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

	videoIDs := utils.MatchOneOf(url, `/video/(\d+)`)
	if len(videoIDs) == 0 {
		return nil, errors.New("unable to get video ID")
	}
	videoID := videoIDs[1]

	dytks := utils.MatchOneOf(html, `dytk: "(.+?)"`)
	if len(dytks) == 0 {
		return nil, errors.New("unable to get dytk info")
	}
	dytk := dytks[1]

	apiDataString, err := request.Get(
		fmt.Sprintf("https://www.douyin.com/web/api/v2/aweme/iteminfo/?item_ids=%s&dytk=%s", videoID, dytk),
		url, nil,
	)
	if err != nil {
		return nil, err
	}

	var apiData data
	if err = json.Unmarshal([]byte(apiDataString), &apiData); err != nil {
		return nil, err
	}

	return []*types.Data{
		{
			Site:    "抖音 douyin.com",
			Title:   apiData.ItemList[0].Desc,
			Type:    types.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
