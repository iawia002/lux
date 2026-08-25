package twitter

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iawia002/lux/extractors"
)

func TestExtractWithXquik(t *testing.T) {
	const (
		tweetID  = "1742731788961853831"
		tweetURL = "https://x.com/fasc1nate/status/1742731788961853831"
	)
	t.Setenv(xquikAPIKeyEnv, "test-api-key")

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/x/tweets/" + tweetID:
			if got := r.Header.Get("x-api-key"); got != "test-api-key" {
				t.Errorf("x-api-key = %q, want test-api-key", got)
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{
				"tweet": {
					"id": %q,
					"media": [
						{"type":"photo","mediaUrl":%q,"url":%q},
						{"type":"video","videoVariants":[
							{"bitrate":832000,"contentType":"video/mp4","url":%q},
							{"bitrate":2176000,"contentType":"video/mp4; codecs=avc1","url":%q},
							{"bitrate":2176000,"contentType":"video/mp4","url":%q},
							{"contentType":"application/x-mpegURL","url":%q}
						]},
						{"type":"animated_gif","videoVariants":[
							{"contentType":"video/mp4","url":%q}
						]}
					]
				},
				"author":{"name":"Fascinating","username":"fasc1nate"}
			}`,
				tweetID,
				server.URL+"/photo.jpg",
				server.URL+"/photo",
				server.URL+"/video-832.mp4",
				server.URL+"/video-2176.mp4",
				server.URL+"/video-2176.mp4",
				server.URL+"/video.m3u8",
				server.URL+"/animation.mp4",
			)
		case "/video-832.mp4":
			serveMedia(t, w, r, tweetURL, "small")
		case "/video-2176.mp4":
			serveMedia(t, w, r, tweetURL, "larger")
		case "/animation.mp4":
			serveMedia(t, w, r, tweetURL, "gif")
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	e := &extractor{
		xquikEndpoint: server.URL + "/api/v1/x/tweets/%s",
		client:        server.Client(),
	}
	data, err := e.Extract(tweetURL, extractors.Options{})
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}
	if got, want := len(data), 2; got != want {
		t.Fatalf("len(data) = %d, want %d", got, want)
	}
	if got, want := data[0].Title, "Fascinating (@fasc1nate) "+tweetID+" [1]"; got != want {
		t.Errorf("first title = %q, want %q", got, want)
	}
	if got, want := data[1].Title, "Fascinating (@fasc1nate) "+tweetID+" [2]"; got != want {
		t.Errorf("second title = %q, want %q", got, want)
	}
	if got, want := data[0].Site, "X (Twitter) x.com"; got != want {
		t.Errorf("site = %q, want %q", got, want)
	}
	if got, want := len(data[0].Streams), 2; got != want {
		t.Fatalf("len(first streams) = %d, want %d", got, want)
	}
	if got, want := data[0].Streams["832000"].Quality, "832 kbps MP4"; got != want {
		t.Errorf("quality = %q, want %q", got, want)
	}
	if got, want := data[0].Streams["2176000"].Size, int64(len("larger")); got != want {
		t.Errorf("size = %d, want %d", got, want)
	}
	if got, want := data[1].Streams["default"].Parts[0].URL, server.URL+"/animation.mp4"; got != want {
		t.Errorf("animation URL = %q, want %q", got, want)
	}
}

func TestExtractWithoutXquikUsesTwitterEndpoint(t *testing.T) {
	const tweetID = "898217160060698624"
	t.Setenv(xquikAPIKeyEnv, "")

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/status/" + tweetID:
			fmt.Fprint(w, `<html><title>X</title></html>`)
		case "/twitter/" + tweetID:
			fmt.Fprintf(w, `{"track":{"playbackUrl":%q}}`, server.URL+"/video.mp4")
		case "/video.mp4":
			serveMedia(t, w, r, server.URL+"/status/"+tweetID, "video")
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	e := &extractor{twitterEndpoint: server.URL + "/twitter/%s"}
	data, err := e.Extract(server.URL+"/status/"+tweetID, extractors.Options{})
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}
	if got, want := len(data), 1; got != want {
		t.Fatalf("len(data) = %d, want %d", got, want)
	}
	if got, want := data[0].Title, "Twitter "+tweetID; got != want {
		t.Errorf("title = %q, want %q", got, want)
	}
	if got, want := data[0].Streams["default"].Size, int64(len("video")); got != want {
		t.Errorf("size = %d, want %d", got, want)
	}
}

func TestTwitterTitle(t *testing.T) {
	tests := []struct {
		name string
		html string
		uri  string
		want string
	}{
		{
			name: "Open Graph title",
			html: `<meta property="og:title" content="Justin Bieber on Twitter">`,
			uri:  "https://twitter.com/justinbieber/status/898217160060698624",
			want: "Justin Bieber on Twitter",
		},
		{
			name: "username fallback",
			uri:  "https://twitter.com/Chaosxsilencer/status/1591088609829847040",
			want: "@Chaosxsilencer",
		},
		{
			name: "generic fallback",
			uri:  "https://twitter.com/i/status/1742731788961853831",
			want: "Twitter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := twitterTitle(tt.html, tt.uri); got != tt.want {
				t.Errorf("twitterTitle() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractWithXquikRejectsInvalidResponse(t *testing.T) {
	t.Setenv(xquikAPIKeyEnv, "test-api-key")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "not JSON")
	}))
	t.Cleanup(server.Close)

	e := &extractor{
		xquikEndpoint: server.URL + "/api/v1/x/tweets/%s",
		client:        server.Client(),
	}
	_, err := e.Extract("https://x.com/user/status/1742731788961853831", extractors.Options{})
	if !errors.Is(err, extractors.ErrBodyParseFailed) {
		t.Fatalf("Extract() error = %v, want ErrBodyParseFailed", err)
	}
}

func TestExtractWithXquikRequiresMP4Video(t *testing.T) {
	t.Setenv(xquikAPIKeyEnv, "test-api-key")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"tweet":{"media":[{"type":"photo","mediaUrl":"https://example.com/photo.jpg","url":"https://example.com/photo"}]}}`)
	}))
	t.Cleanup(server.Close)

	e := &extractor{
		xquikEndpoint: server.URL + "/api/v1/x/tweets/%s",
		client:        server.Client(),
	}
	_, err := e.Extract("https://x.com/user/status/1742731788961853831", extractors.Options{})
	if err == nil || err.Error() != "tweet has no downloadable MP4 video" {
		t.Fatalf("Extract() error = %v, want no downloadable video error", err)
	}
}

func TestXquikHTTPErrorDoesNotExposeAPIKey(t *testing.T) {
	const apiKey = "test-api-key"
	t.Setenv(xquikAPIKeyEnv, apiKey)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	e := &extractor{
		xquikEndpoint: server.URL + "/api/v1/x/tweets/%s",
		client:        server.Client(),
	}
	_, err := e.Extract("https://x.com/user/status/1742731788961853831", extractors.Options{})
	if err == nil {
		t.Fatal("Extract() error = nil, want HTTP error")
	}
	if strings.Contains(err.Error(), apiKey) {
		t.Fatalf("Extract() exposed the API key: %v", err)
	}
}

func TestExtractRejectsURLWithoutTweetID(t *testing.T) {
	t.Setenv(xquikAPIKeyEnv, "test-api-key")
	e := &extractor{xquikEndpoint: "https://example.com/%s"}
	_, err := e.Extract("https://x.com/user", extractors.Options{})
	if !errors.Is(err, extractors.ErrURLParseFailed) {
		t.Fatalf("Extract() error = %v, want ErrURLParseFailed", err)
	}
}

func TestXDotComUsesTwitterExtractor(t *testing.T) {
	_, err := extractors.Extract("https://x.com/user", extractors.Options{})
	if !errors.Is(err, extractors.ErrURLParseFailed) {
		t.Fatalf("Extract() error = %v, want ErrURLParseFailed", err)
	}
}

func serveMedia(t *testing.T, w http.ResponseWriter, r *http.Request, referer, body string) {
	t.Helper()
	if got := r.Header.Get("Referer"); got != referer {
		t.Errorf("Referer = %q, want %q", got, referer)
	}
	w.Header().Set("Content-Length", fmt.Sprint(len(body)))
	fmt.Fprint(w, body)
}
