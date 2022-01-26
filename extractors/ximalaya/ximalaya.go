package ximalaya

import (
	"encoding/json"
	"errors"

	"github.com/iawia002/lux/extractors/types"
	"github.com/iawia002/lux/parser"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

type extractor struct{}

// New returns a ximalaya extractor.
func New() types.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option types.Options) ([]*types.Data, error) {
	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, err
	}

	// get the title
	doc, err := parser.GetDoc(html)
	if err != nil {
		return nil, err
	}
	title := parser.Title(doc)

	itemIds := utils.MatchOneOf(url, `/sound/(\d+)`)
	if len(itemIds) == 0 {
		return nil, errors.New("unable to get audio ID")
	}
	if itemIds == nil || len(itemIds) < 2 {
		return nil, types.ErrURLParseFailed
	}
	itemId := itemIds[len(itemIds)-1]

	jsonData, err := request.Get("https://www.ximalaya.com/revision/play/v1/audio?id="+itemId+"&ptype=1", url, nil)
	if err != nil {
		return nil, err
	}
	var ximalaya ximalayaData
	if err = json.Unmarshal([]byte(jsonData), &ximalaya); err != nil {
		return nil, err
	}

	realURL := ximalaya.Data.Src
	urlData := make([]*types.Part, 0)
	totalSize, err := request.Size(realURL, url)
	if err != nil {
		return nil, err
	}
	_, ext, err := utils.GetNameAndExt(realURL)
	if err != nil {
		return nil, err
	}
	urlData = append(urlData, &types.Part{
		URL:  realURL,
		Size: totalSize,
		Ext:  ext,
	})
	streams := map[string]*types.Stream{
		"default": {
			Parts: urlData,
			Size:  totalSize,
		},
	}

	return []*types.Data{
		{
			Site:    "喜马拉雅 ximalaya.com",
			Title:   title,
			Type:    types.DataTypeAudio,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
