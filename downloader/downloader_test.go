package downloader

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/iawia002/lux/extractors"
)

func TestDownload(t *testing.T) {
	// Use a local test server instead of external URLs that may become unavailable.
	// See https://github.com/iawia002/lux/issues/1413
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/video.mp4":
			w.Header().Set("Content-Type", "video/mp4")
			w.Header().Set("Content-Length", "12")
			w.Write([]byte("fake-video!!")) // nolint
		case "/image1.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			w.Header().Set("Content-Length", "10")
			w.Write([]byte("fake-img-1")) // nolint
		case "/image2.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			w.Header().Set("Content-Length", "10")
			w.Write([]byte("fake-img-2")) // nolint
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	testCases := []struct {
		name string
		data *extractors.Data
	}{
		{
			name: "normal test",
			data: &extractors.Data{
				Site:  "test",
				Title: "test",
				Type:  extractors.DataTypeVideo,
				URL:   ts.URL,
				Streams: map[string]*extractors.Stream{
					"default": {
						ID: "default",
						Parts: []*extractors.Part{
							{
								URL:  ts.URL + "/video.mp4",
								Size: 12,
								Ext:  "mp4",
							},
						},
						Size: 12,
					},
				},
			},
		},
		{
			name: "multi-stream test",
			data: &extractors.Data{
				Site:  "test",
				Title: "test2",
				Type:  extractors.DataTypeVideo,
				URL:   ts.URL,
				Streams: map[string]*extractors.Stream{
					"stream-a": {
						ID: "stream-a",
						Parts: []*extractors.Part{
							{
								URL:  ts.URL + "/video.mp4",
								Size: 12,
								Ext:  "mp4",
							},
						},
						Size: 12,
					},
					"stream-b": {
						ID: "stream-b",
						Parts: []*extractors.Part{
							{
								URL:  ts.URL + "/video.mp4",
								Size: 12,
								Ext:  "mp4",
							},
						},
						Size: 12,
					},
				},
			},
		},
		{
			name: "image test",
			data: &extractors.Data{
				Site:  "test",
				Title: "test-image",
				Type:  extractors.DataTypeImage,
				URL:   ts.URL,
				Streams: map[string]*extractors.Stream{
					"default": {
						ID: "default",
						Parts: []*extractors.Part{
							{
								URL:  ts.URL + "/image1.jpg",
								Size: 10,
								Ext:  "jpg",
							},
							{
								URL:  ts.URL + "/image2.jpg",
								Size: 10,
								Ext:  "jpg",
							},
						},
					},
				},
			},
		},
	}

	// Use a temp directory for downloads
	tmpDir, err := os.MkdirTemp("", "lux-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := New(Options{Silent: true, OutputPath: tmpDir}).Download(testCase.data)
			if err != nil {
				t.Errorf("%s: %v", testCase.name, err)
			}
		})
	}
}
