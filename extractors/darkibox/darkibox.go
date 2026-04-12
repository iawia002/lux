package darkibox

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/robertkrimen/otto"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("darkibox", New())
}

type extractor struct{}

// New returns a darkibox extractor.
func New() extractors.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	fileCode, err := extractFileCode(url)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	embedURL := fmt.Sprintf("https://darkibox.com/embed-%s.html", fileCode)

	// Get the embed page to extract the title
	html, err := request.Get(embedURL, url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	title := "Darkibox Video"
	titleMatch := utils.MatchOneOf(html, `<title>([^<]+)</title>`)
	if len(titleMatch) >= 2 {
		t := strings.TrimSpace(titleMatch[1])
		if t != "" {
			title = t
		}
	}

	// POST to /dl endpoint to get the packed JS with video URL
	formData := fmt.Sprintf("op=embed&file_code=%s&auto=1", fileCode)
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"Referer":      embedURL,
	}
	res, err := request.Request(http.MethodPost, "https://darkibox.com/dl", strings.NewReader(formData), headers)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer res.Body.Close() // nolint

	bodyBytes, err := readResponseBody(res)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	body := string(bodyBytes)

	// Extract and unpack the packed JavaScript
	videoURL, err := extractVideoURL(body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	streams := make(map[string]*extractors.Stream)

	// Check if the URL is an m3u8 (HLS) stream
	if strings.Contains(videoURL, ".m3u8") {
		streams["default"] = &extractors.Stream{
			Parts: []*extractors.Part{
				{
					URL:  videoURL,
					Size: 0,
					Ext:  "ts",
				},
			},
		}
	} else {
		// Direct MP4 or other format
		size, _ := request.Size(videoURL, embedURL)
		streams["default"] = &extractors.Stream{
			Parts: []*extractors.Part{
				{
					URL:  videoURL,
					Size: size,
					Ext:  "mp4",
				},
			},
			Size: size,
		}
	}

	return []*extractors.Data{
		{
			Site:    "Darkibox darkibox.com",
			Title:   title,
			Type:    extractors.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}

// extractFileCode extracts the file code from various darkibox URL formats:
//   - darkibox.com/FILECODE
//   - darkibox.com/d/FILECODE
//   - darkibox.com/embed-FILECODE.html
func extractFileCode(url string) (string, error) {
	// Try embed format first
	match := utils.MatchOneOf(url, `darkibox\.com/embed-([a-zA-Z0-9]+)\.html`)
	if len(match) >= 2 {
		return match[1], nil
	}

	// Try /d/FILECODE format
	match = utils.MatchOneOf(url, `darkibox\.com/d/([a-zA-Z0-9]+)`)
	if len(match) >= 2 {
		return match[1], nil
	}

	// Try direct FILECODE format (darkibox.com/FILECODE)
	match = utils.MatchOneOf(url, `darkibox\.com/([a-zA-Z0-9]+)`)
	if len(match) >= 2 {
		return match[1], nil
	}

	return "", errors.New("could not extract file code from URL")
}

// extractVideoURL extracts the video URL from the response body.
// It first tries to unpack packed JavaScript, then falls back to regex.
func extractVideoURL(body string) (string, error) {
	// Try to find and unpack eval(function(p,a,c,k,e,d){...}) packed JS
	packedMatch := utils.MatchOneOf(body, `(eval\(function\(p,a,c,k,e,d\)\{.*?\}\(.*?\)\))`)
	if len(packedMatch) >= 2 {
		unpacked, err := unpackJS(packedMatch[1])
		if err == nil {
			// Look for file:"URL" in unpacked JS (PlayerJS format)
			videoURL := extractFileURL(unpacked)
			if videoURL != "" {
				return videoURL, nil
			}
		}
	}

	// Fallback: try to find video URL directly in the body
	videoURL := extractFileURL(body)
	if videoURL != "" {
		return videoURL, nil
	}

	// Try to find any m3u8 or mp4 URL
	urlMatch := utils.MatchOneOf(body, `(https?://[^\s"']+\.(?:m3u8|mp4)[^\s"']*)`)
	if len(urlMatch) >= 2 {
		return urlMatch[1], nil
	}

	return "", errors.New("could not extract video URL from response")
}

// extractFileURL extracts a URL from file:"URL" or file: "URL" patterns (PlayerJS).
func extractFileURL(s string) string {
	// Handle multi-quality format: [label]url,[label]url
	match := utils.MatchOneOf(s, `file\s*:\s*"((?:\[[^\]]+\])?https?://[^"]+)"`)
	if len(match) >= 2 {
		fileValue := match[1]
		// If it contains [label] multi-quality format, extract the highest quality
		if strings.Contains(fileValue, "[") {
			return extractBestQuality(fileValue)
		}
		return fileValue
	}
	return ""
}

// extractBestQuality extracts the best quality URL from a multi-quality string
// format: [1080p]https://url1.m3u8,[720p]https://url2.m3u8,...
func extractBestQuality(s string) string {
	re := regexp.MustCompile(`\[([^\]]+)\](https?://[^,\s"]+)`)
	matches := re.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return ""
	}
	// Return the first one (typically highest quality)
	return matches[0][2]
}

// unpackJS unpacks eval(function(p,a,c,k,e,d){...}) packed JavaScript using otto.
func unpackJS(packed string) (string, error) {
	vm := otto.New()

	// Replace eval with a function that captures the output
	script := fmt.Sprintf(`var __unpacked = ""; function eval(s) { __unpacked = s; return s; }; %s;`, packed)
	_, err := vm.Run(script)
	if err != nil {
		return "", errors.WithStack(err)
	}

	value, err := vm.Get("__unpacked")
	if err != nil {
		return "", errors.WithStack(err)
	}

	result, err := value.ToString()
	if err != nil {
		return "", errors.WithStack(err)
	}

	if result == "" || result == "undefined" {
		return "", errors.New("unpacking returned empty result")
	}

	return result, nil
}

// readResponseBody reads the full response body.
func readResponseBody(res *http.Response) ([]byte, error) {
	return io.ReadAll(res.Body)
}

