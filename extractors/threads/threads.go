package threads

import (
	"fmt"
	"net"
	"net/http"
	netURL "net/url"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("threads", New())
}

type extractor struct {
	client *http.Client
}

// New returns a instagram extractor.
func New() extractors.Extractor {
	return &extractor{
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).Dial,
				TLSHandshakeTimeout: 5 * time.Second,
			},
		},
	}
}

type media struct {
	URL  string
	Type extractors.DataType
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	URL, err := netURL.Parse(url)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	paths := strings.Split(URL.Path, "/")
	if len(paths) < 3 {
		return nil, errors.New("invalid URL format")
	}

	poster := paths[1]
	shortCode := paths[3]

	medias := make([]media, 0)

	title := fmt.Sprintf("Threads %s - %s", poster, shortCode)

	collector := colly.NewCollector()
	collector.SetClient(e.client)

	// case single image or video
	collector.OnHTML("div.SingleInnerMediaContainer", func(e *colly.HTMLElement) {
		if src := e.ChildAttr("img", "src"); src != "" {
			medias = append(medias, media{
				URL:  src,
				Type: extractors.DataTypeImage,
			})
		}
		if src := e.ChildAttr("video > source", "src"); src != "" {
			medias = append(medias, media{
				URL:  src,
				Type: extractors.DataTypeVideo,
			})
		}
	})

	// case multiple image or video
	collector.OnHTML("div.MediaScrollImageContainer", func(e *colly.HTMLElement) {
		if src := e.ChildAttr("img", "src"); src != "" {
			medias = append(medias, media{
				URL:  src,
				Type: extractors.DataTypeImage,
			})
		}
		if src := e.ChildAttr("video > source", "src"); src != "" {
			medias = append(medias, media{
				URL:  src,
				Type: extractors.DataTypeVideo,
			})
		}
	})

	// title with caption
	// collector.OnHTML("span.BodyTextContainer", func(e *colly.HTMLElement) {
	// 	title = e.Text
	// })

	if err := collector.Visit(URL.JoinPath("embed").String()); err != nil {
		return nil, fmt.Errorf("failed to send HTTP request to the Threads: %w", errors.WithStack(err))
	}

	var totalSize int64
	var parts []*extractors.Part

	for _, m := range medias {
		_, ext, err := utils.GetNameAndExt(m.URL)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		fileSize, err := request.Size(m.URL, url)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		part := &extractors.Part{
			URL:  m.URL,
			Size: fileSize,
			Ext:  ext,
		}
		parts = append(parts, part)
	}

	for _, part := range parts {
		totalSize += part.Size
	}

	streams := map[string]*extractors.Stream{
		"default": {
			Parts: parts,
			Size:  totalSize,
		},
	}

	return []*extractors.Data{
		{
			Site:    "Threads www.threads.net",
			Title:   title,
			Type:    extractors.DataTypeImage,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
