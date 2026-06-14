package extractors

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OldBeLight/lux/request"
)

type libraryExtractor struct {
	serverURL string
}

func (e libraryExtractor) Extract(u string, option Options) ([]*Data, error) {
	body, err := request.Get(e.serverURL, u, nil)
	if err != nil {
		return nil, err
	}
	return []*Data{{
		URL:   u,
		Site:  body,
		Title: "library",
		Type:  DataTypeVideo,
		Streams: map[string]*Stream{
			"default": {
				Parts: []*Part{{URL: e.serverURL, Size: 1, Ext: "mp4"}},
			},
		},
	}}, nil
}

func TestExtractWithRequestOptionsScopesRequestOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "library-agent" {
			t.Fatalf("User-Agent = %q, want library-agent", r.Header.Get("User-Agent"))
		}
		if cookie, err := r.Cookie("session"); err != nil || cookie.Value != "abc" {
			t.Fatalf("cookie = %#v, %v; want session=abc", cookie, err)
		}
		_, _ = w.Write([]byte("library-site"))
	}))
	defer server.Close()

	Register("example", libraryExtractor{serverURL: server.URL})
	videos, err := ExtractWithRequestOptions("http://example.com/watch", Options{}, request.Options{
		RetryTimes: 1,
		Cookie:     "session=abc",
		UserAgent:  "library-agent",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(videos) != 1 || videos[0].Site != "library-site" {
		t.Fatalf("videos = %#v", videos)
	}
}
