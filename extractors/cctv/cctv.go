package cctv

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"github.com/wujiu2020/lux/extractors/proto"
	"github.com/wujiu2020/lux/request"
)

const (
	api = "https://vdn.apps.cntv.cn/api/getHttpVideoInfo.do?pid=%s"
)

func New() proto.Extractor {
	return &extractor{}
}

type extractor struct{}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string) (proto.TransformData, error) {
	html, err := request.Get(url, "", nil)
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`var guid = "(.*?)";`)
	match := re.FindStringSubmatch(html)
	if len(match) == 0 {
		return nil, errors.New("have no mathc url")
	}
	pid := match[1]
	resByte, err := request.GetByte(fmt.Sprintf(api, pid), "", nil)
	if err != nil {
		return nil, err
	}
	var videoInfoRes VideoInfoRes
	if err := json.Unmarshal(resByte, &videoInfoRes); err != nil {
		return nil, err
	}
	return videoInfoRes, nil
}
