package alipay

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iawia002/lux/extractors"
)

func TestContentID(t *testing.T) {
	for _, text := range []string{
		"https://example.com/?contentId=ABC123&other=1",
		"alipays://start?url=%2Fdetail.html%3FcontentId%3DABC123%26other%3D1",
		"contentId%253DABC123",
	} {
		if got := contentID(text); got != "ABC123" {
			t.Fatalf("contentID(%q) = %q", text, got)
		}
	}
	if got := contentID("https://example.com/no-video"); got != "" {
		t.Fatal(got)
	}
}

func TestStreamURLs(t *testing.T) {
	streams, err := streamURLs("https://cdn.example/video/720P_h265?t=token&extra=value")
	if err != nil || len(streams) != len(clarities) {
		t.Fatalf("%v: %v", streams, err)
	}
	if streams["1080P_h265"] != "https://cdn.example/video/1080P_h265?t=token&extra=value" {
		t.Fatal(streams)
	}
	streams, err = streamURLs("https://cdn.example/video.mp4")
	if err != nil || len(streams) != 1 {
		t.Fatalf("%v: %v", streams, err)
	}
	if _, err = streamURLs("file:///tmp/video.mp4"); err == nil {
		t.Fatal("expected URL error")
	}
}

func TestExtract(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/share":
			fmt.Fprint(w, `alipays://start?url=%2Fdetail%3FcontentId%3DABC123`)
		case "/redirect":
			http.Redirect(w, r, "/detail?contentId=ABC123", http.StatusFound)
		case "/detail":
		case "/api":
			if r.Method != http.MethodPost || r.Header.Get("x-webgw-appId") != "180020010001266490" || r.Header.Get("x-webgw-version") != "2.0" || r.Header.Get("Content-Type") != "application/json" {
				t.Error("invalid API request")
			}
			var payload map[string]string
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload["contentId"] != "ABC123" {
				t.Error("invalid payload")
			}
			w.Header().Set("Content-Encoding", "gzip")
			compressed := gzip.NewWriter(w)
			fmt.Fprintf(compressed, `{"success":true,"resultObj":{"data":[{"title":"test title","video":{"vid":%q}}]}}`, server.URL+"/video/720P_h265?t=token&extra=value")
			compressed.Close()
		case "/video/1080P_h265":
			if r.Header.Get("Range") != "bytes=0-0" || r.URL.Query().Get("t") != "token" || r.URL.Query().Get("extra") != "value" {
				t.Error("invalid probe")
			}
			w.Header().Set("Content-Range", "bytes 0-0/12345")
			w.WriteHeader(http.StatusPartialContent)
			fmt.Fprint(w, "x")
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	e := &extractor{request: func(method, link string, body io.Reader, headers map[string]string) (*http.Response, error) {
		if link == webGW {
			link = server.URL + "/api"
		}
		req, err := http.NewRequest(method, link, body)
		if err != nil {
			return nil, err
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		return server.Client().Do(req)
	}}
	for _, path := range []string{"/share", "/redirect", "/detail?contentId=ABC123"} {
		data, err := e.Extract(server.URL+path, extractors.Options{})
		if err != nil {
			t.Fatal(err)
		}
		if data[0].Title != "test title" || len(data[0].Streams) != 1 || data[0].Streams["1080P_h265"].Size != 12345 {
			t.Fatalf("unexpected result: %+v", data[0])
		}
	}
	if _, err := e.resolveID(server.URL + "/detail"); err == nil {
		t.Fatal("expected missing ID error")
	}
}

func TestAPIErrors(t *testing.T) {
	for _, body := range []string{
		`not json`, `{"success":false,"errorMessage":"gone"}`,
		`{"success":true,"resultObj":{"data":[]}}`,
		`{"success":true,"resultObj":{"data":[{"video":null}]}}`,
	} {
		e := &extractor{request: func(string, string, io.Reader, map[string]string) (*http.Response, error) {
			return &http.Response{Body: io.NopCloser(strings.NewReader(body))}, nil
		}}
		if _, err := e.Extract("https://example.com/?contentId=ABC123", extractors.Options{}); err == nil {
			t.Fatalf("expected error for %s", body)
		}
	}
}

func TestStreamSize(t *testing.T) {
	for _, test := range []struct {
		status  int
		headers http.Header
		want    int64
	}{
		{200, http.Header{"Content-Length": {"42"}}, 42},
		{206, http.Header{"Content-Length": {"1"}, "Content-Range": {"bytes 0-0/100"}}, 100},
		{206, http.Header{"Content-Length": {"1"}}, 0},
		{404, http.Header{}, 0},
	} {
		e := &extractor{request: func(string, string, io.Reader, map[string]string) (*http.Response, error) {
			return &http.Response{StatusCode: test.status, Header: test.headers, Body: io.NopCloser(strings.NewReader("x"))}, nil
		}}
		got, err := e.streamSize("https://example.com/video")
		if got != test.want || (err != nil) != (test.want == 0) {
			t.Fatalf("got %d, %v; want %d", got, err, test.want)
		}
	}
}
