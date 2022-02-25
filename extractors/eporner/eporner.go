package eporner

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/parser"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("eporner", New())
}

const (
	downloadclass = ".dloaddivcol"
)

type src struct {
	url     string
	quality string
	sizestr string
	size    int64
}

func getSrcMeta(text string) *src {
	sti := strings.Index(text, "(")
	ste := strings.Index(text, ")")
	itext := text[sti+1 : ste]
	strs := strings.Split(itext, ",")
	s := &src{}

	if len(strs) == 2 {
		s.quality = strings.Trim(strs[0], " ")
		s.sizestr = strings.Trim(strs[1], " ")
	}

	if s.sizestr == "" {
		s.size = 0
		return s
	}

	valunit := strings.Split(s.sizestr, " ")
	val, err := strconv.ParseFloat(valunit[0], 64)
	if err != nil {
		s.size = 0
		return s
	}
	unit := valunit[1]
	switch unit {
	case "KB":
		s.size = int64(val * 1024)
	case "MB":
		s.size = int64(val * 1024 * 1024)
	case "GB":
		s.size = int64(val * 1024 * 1024 * 1024)
	default:
		s.size = int64(val)
	}
	return s
}

func getSrc(html string) []*src {
	srcs := []*src{}
	d, err := parser.GetDoc(html)
	if err != nil {
		return nil
	}

	d.Find(downloadclass).Each(func(i int, s *goquery.Selection) {
		s.Contents().Each(func(i int, s *goquery.Selection) {
			for ns := range s.Nodes {
				n := s.Get(ns)
				if n.Data == "a" {
					var sr *src
					if n.FirstChild != nil {
						sr = getSrcMeta(n.FirstChild.Data)
					}
					for _, a := range n.Attr {
						if a.Key == "href" {
							sr.url = a.Val
						}
					}
					srcs = append(srcs, sr)
				}
			}
		})
	})

	return srcs
}

type extractor struct{}

// New returns a eporner extractor.
func New() extractors.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(u string, option extractors.Options) ([]*extractors.Data, error) {
	html, err := request.Get(u, u, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var title string
	desc := utils.MatchOneOf(html, `<title>(.+?)</title>`)
	if len(desc) > 1 {
		title = desc[1]
	} else {
		title = "eporner"
	}
	uu, err := url.Parse(u)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	srcs := getSrc(html)
	streams := make(map[string]*extractors.Stream, len(srcs))
	for _, src := range srcs {
		srcurl := uu.Scheme + "://" + uu.Host + src.url
		// skipping an extra HEAD request to the URL.
		// size, err := request.Size(srcurl, u)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		urlData := &extractors.Part{
			URL:  srcurl,
			Size: src.size,
			Ext:  "mp4",
		}
		streams[src.quality] = &extractors.Stream{
			Parts:   []*extractors.Part{urlData},
			Size:    src.size,
			Quality: src.quality,
		}
	}
	return []*extractors.Data{
		{
			Site:    "EPORNER eporner.com",
			Title:   title,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     u,
		},
	}, nil
}
