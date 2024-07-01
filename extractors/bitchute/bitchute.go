package bitchute

import (
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
)

func init() {
	extractors.Register("bitchute", New())
}

type extractor struct{}

// New returns a bitchute extractor.
func New() extractors.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(u string, option extractors.Options) ([]*extractors.Data, error) {
	regVideoID := regexp.MustCompile(`/video/([^?/]+)`)
	matchVideoID := regVideoID.FindStringSubmatch(u)
	if len(matchVideoID) < 2 {
		return nil, errors.New("Invalid video URL: Missing video ID parameter")
	}
	embedURL := fmt.Sprintf("https://www.bitchute.com/api/beta9/embed/?videoID=%s", matchVideoID[1])

	res, err := request.Request(http.MethodGet, embedURL, nil, nil)
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

	// There is also an API that provides meta data
	// POST https://api.bitchute.com/api/beta9/video {"video_id": <video_id>}
	html := string(b)
	regMediaURL := regexp.MustCompile(`media_url\s*=\s*['|"](https:\/\/[^.]+\.bitchute\.com\/.*\.mp4)`)
	matchMediaURL := regMediaURL.FindStringSubmatch(html)
	if len(matchMediaURL) < 2 {
		return nil, errors.New("Could not extract media URL from page")
	}
	mediaURL := matchMediaURL[1]

	regVideoName := regexp.MustCompile(`(?m)video_name\s*=\s*["|']\\?"?(.*)["|'];?$`)
	matchVideoName := regVideoName.FindStringSubmatch(html)
	if len(matchVideoName) < 2 {
		return nil, errors.New("Could not extract media name from page")
	}
	videoName := strings.ReplaceAll(matchVideoName[1], `\"`, "")

	streams := make(map[string]*extractors.Stream, 1)
	size, err := request.Size(mediaURL, u)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	streams["Default"] = &extractors.Stream{
		Parts: []*extractors.Part{
			{
				URL:  mediaURL,
				Size: size,
				Ext:  "mp4",
			},
		},
		Size:    size,
		Quality: "Default",
	}

	return []*extractors.Data{
		{
			Site:    "Bitchute bitchute.com",
			Title:   videoName,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     u,
		},
	}, nil
}
