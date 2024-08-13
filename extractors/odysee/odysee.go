package odysee

import (
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"regexp"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/pkg/errors"
)

func init() {
	extractors.Register("odysee", New())
}

type extractor struct{}

type odyseePayload struct {
	ContentURL   string `json:"contentUrl"`
	Description  string `json:"description"`
	Name         string `json:"name"`
	ThumbnailURL string `json:"thumbnailUrl"`
	URL          string `json:"url"`
}

// New returns an odysee extractor.
func New() extractors.Extractor {
	return &extractor{}
}

func (e *extractor) Extract(u string, option extractors.Options) ([]*extractors.Data, error) {
	res, err := request.Request(http.MethodGet, u, nil, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer res.Body.Close() // nolint

	var reader io.ReadCloser
	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		reader, _ = gzip.NewReader(res.Body)
	case "deflate":
		reader = flate.NewReader(res.Body)
	default:
		reader = res.Body
	}
	defer reader.Close() // nolint

	b, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	regScript := regexp.MustCompile(`(?im)\<script type="application\/ld\+json"\>([\s\S]*)[\n?]<\/script>`)
	matchPayload := regScript.FindSubmatch(b)
	if len(matchPayload) < 2 {
		return nil, errors.New("Could not read page data")
	}

	var resData odyseePayload
	if err := json.Unmarshal(matchPayload[1], &resData); err != nil {
		return nil, errors.WithStack(err)
	}

	streams := make(map[string]*extractors.Stream, 1)
	size, err := request.Size(resData.ContentURL, u)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	streams["Default"] = &extractors.Stream{
		Parts: []*extractors.Part{
			{
				URL:  resData.ContentURL,
				Size: size,
				Ext:  "mp4",
			},
		},
		Size:    size,
		Quality: "Default",
	}

	return []*extractors.Data{
		{
			Site:    "Odysee odysee.com",
			Title:   resData.Name,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     u,
		},
	}, nil
}
