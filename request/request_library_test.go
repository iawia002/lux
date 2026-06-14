package request

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestWithOptionsUsesProxy(t *testing.T) {
	var proxied bool
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxied = true
		if r.URL.String() != "http://example.test/video" {
			t.Fatalf("proxy saw URL %q, want absolute target URL", r.URL.String())
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok")
	}))
	defer proxy.Close()

	res, err := RequestWithOptions(http.MethodGet, "http://example.test/video", nil, nil, Options{
		RetryTimes: 1,
		Proxy:      proxy.URL,
	})
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()

	if !proxied {
		t.Fatal("request did not go through proxy")
	}
}

func TestWithOptionsRestoresPreviousOptions(t *testing.T) {
	SetOptions(Options{RetryTimes: 3, Cookie: "a=b", UserAgent: "before", Refer: "https://before.example", Proxy: "http://proxy-before", Debug: true})

	err := WithOptions(Options{RetryTimes: 1, UserAgent: "inside"}, func() error {
		optionsMu.Lock()
		defer optionsMu.Unlock()
		current := currentOptionsUnlocked()
		if current.UserAgent != "inside" || current.RetryTimes != 1 {
			t.Fatalf("scoped options = %#v", current)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	optionsMu.Lock()
	defer optionsMu.Unlock()
	current := currentOptionsUnlocked()
	if current.UserAgent != "before" || current.Cookie != "a=b" || current.Proxy != "http://proxy-before" || !current.Debug {
		t.Fatalf("options were not restored: %#v", current)
	}
}
