package streamtape

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/robertkrimen/otto"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	e := New()
	extractors.Register("streamtape", e)
	extractors.Register("streamta", e) // streamta.pe
}

type extractor struct{}

// New returns a StreamTape extractor
func New() extractors.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, _ extractors.Options) ([]*extractors.Data, error) {
	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	scripts := utils.MatchOneOf(html, `document.getElementById\('norobotlink'\).innerHTML = (.+?);`)
	if len(scripts) < 2 {
		return nil, errors.WithStack(extractors.ErrURLParseFailed)
	}

	vm := otto.New()
	_, err = vm.Run(fmt.Sprintf("var __VM__OUTPUT = %s", scripts[1]))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	value, err := vm.Get("__VM__OUTPUT")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	u, err := value.ToString() // //streamtape.com/get_video?id=xx&expires=xx&ip=xx&token=xx
	if err != nil {
		return nil, errors.WithStack(err)
	}
	u = fmt.Sprintf("https:%s&stream=1", u)

	// get title
	var title = "StreamTape Video"
	titleMatch := utils.MatchOneOf(html,
		`\<meta name="og:title" content="(.*)"\>`)
	if len(titleMatch) >= 2 {
		title = titleMatch[1]
	}

	size, err := request.Size(u, url)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	urlData := &extractors.Part{
		URL:  u,
		Size: size,
		Ext:  "mp4",
	}

	streams := make(map[string]*extractors.Stream)
	streams["default"] = &extractors.Stream{
		Parts: []*extractors.Part{urlData},
		Size:  size,
	}

	return []*extractors.Data{
		{
			URL:     u,
			Site:    "StreamTape streamtape.com",
			Title:   title,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
		},
	}, nil
}
