package bcy

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/parser"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("bcy", New())
}

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

// New returns a bcy extractor.
func New() extractors.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// parse json data
	rep := strings.NewReplacer(`\"`, `"`, `\\`, `\`)
	realURLs := utils.MatchOneOf(html, `JSON.parse\("(.+?)"\);`)
	if realURLs == nil || len(realURLs) < 2 {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}
	jsonString := rep.Replace(realURLs[1])

	var data bcyData
	if err = json.Unmarshal([]byte(jsonString), &data); err != nil {
		return nil, errors.WithStack(err)
	}

	doc, err := parser.GetDoc(html)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	title := strings.Replace(parser.Title(doc), " - 半次元 banciyuan - ACG爱好者社区", "", -1)

	parts := make([]*extractors.Part, 0, len(data.Detail.PostData.Multi))
	var totalSize int64
	for _, img := range data.Detail.PostData.Multi {
		size, err := request.Size(img.OriginalPath, url)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		totalSize += size
		_, ext, err := utils.GetNameAndExt(img.OriginalPath)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		parts = append(parts, &extractors.Part{
			URL:  img.OriginalPath,
			Size: size,
			Ext:  ext,
		})
	}
	streams := map[string]*extractors.Stream{
		"default": {
			Parts: parts,
			Size:  totalSize,
		},
	}
	return []*extractors.Data{
		{
			Site:    "半次元 bcy.net",
			Title:   title,
			Type:    extractors.DataTypeImage,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
