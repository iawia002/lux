package bcy

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/iawia002/annie/extractors/types"
	"github.com/iawia002/annie/parser"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type bcyData struct {
	Detail struct {
		PostData struct {
			Multi []struct {
				OriginalPath string `json:"original_path"`
			} `json:"multi"`
		} `json:"post_data"`
	} `json:"detail"`
}

type extractor struct{}

// New returns a youtube extractor.
func New() types.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option types.Options) ([]*types.Data, error) {
	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, err
	}

	// parse json data
	rep := strings.NewReplacer(`\"`, `"`, `\\`, `\`)
	realURLs := utils.MatchOneOf(html, `JSON.parse\("(.+?)"\);`)
	if realURLs == nil || len(realURLs) < 2 {
		return nil, types.ErrURLParseFailed
	}
	jsonString := rep.Replace(realURLs[1])

	var data bcyData
	if err = json.Unmarshal([]byte(jsonString), &data); err != nil {
		return nil, fmt.Errorf("json unmarshal failed, err: %v", err)
	}

	doc, err := parser.GetDoc(html)
	if err != nil {
		return nil, err
	}
	title := strings.Replace(parser.Title(doc), " - 半次元 banciyuan - ACG爱好者社区", "", -1)

	parts := make([]*types.Part, 0, len(data.Detail.PostData.Multi))
	var totalSize int64
	for _, img := range data.Detail.PostData.Multi {
		size, err := request.Size(img.OriginalPath, url)
		if err != nil {
			return nil, err
		}
		totalSize += size
		_, ext, err := utils.GetNameAndExt(img.OriginalPath)
		if err != nil {
			return nil, err
		}
		parts = append(parts, &types.Part{
			URL:  img.OriginalPath,
			Size: size,
			Ext:  ext,
		})
	}
	streams := map[string]*types.Stream{
		"default": {
			Parts: parts,
			Size:  totalSize,
		},
	}
	return []*types.Data{
		{
			Site:    "半次元 bcy.net",
			Title:   title,
			Type:    types.DataTypeImage,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
