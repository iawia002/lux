package request

import (
	"compress/flate"
	"compress/gzip"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	cookiemonster "github.com/MercuryEngineering/CookieMonster"
	"github.com/fatih/color"
	"github.com/kr/pretty"
	"github.com/pkg/errors"

	"github.com/OldBeLight/lux/config"
)

var (
	retryTimes int
	rawCookie  string
	userAgent  string
	refer      string
	proxy      string
	debug      bool
	ctx        context.Context
	optionsMu  sync.Mutex
	scopedMu   sync.Mutex
)

// Options defines common request options.
type Options struct {
	RetryTimes int
	Context    context.Context
	Cookie     string
	UserAgent  string
	Refer      string
	Proxy      string
	Debug      bool
	Silent     bool
}

// SetOptions sets the common request option.
func SetOptions(opt Options) {
	optionsMu.Lock()
	defer optionsMu.Unlock()
	setOptionsUnlocked(opt)
}

func setOptionsUnlocked(opt Options) {
	retryTimes = opt.RetryTimes
	rawCookie = opt.Cookie
	userAgent = opt.UserAgent
	refer = opt.Refer
	proxy = opt.Proxy
	debug = opt.Debug
	ctx = opt.Context
}

func currentOptionsUnlocked() Options {
	return Options{
		RetryTimes: retryTimes,
		Context:    ctx,
		Cookie:     rawCookie,
		UserAgent:  userAgent,
		Refer:      refer,
		Proxy:      proxy,
		Debug:      debug,
	}
}

// CurrentOptions returns a snapshot of the current request options.
func CurrentOptions() Options {
	optionsMu.Lock()
	defer optionsMu.Unlock()
	return currentOptionsUnlocked()
}

// WithOptions executes fn with request options scoped to this library call.
func WithOptions(opt Options, fn func() error) error {
	scopedMu.Lock()
	defer scopedMu.Unlock()

	optionsMu.Lock()
	previous := currentOptionsUnlocked()
	setOptionsUnlocked(opt)
	optionsMu.Unlock()

	defer func() {
		optionsMu.Lock()
		setOptionsUnlocked(previous)
		optionsMu.Unlock()
	}()
	return fn()
}

// Request base request
func Request(method, url string, body io.Reader, headers map[string]string) (*http.Response, error) {
	optionsMu.Lock()
	opt := currentOptionsUnlocked()
	optionsMu.Unlock()
	return RequestWithOptions(method, url, body, headers, opt)
}

// RequestWithOptions runs a request with explicit request options.
func RequestWithOptions(method, url string, body io.Reader, headers map[string]string, opt Options) (*http.Response, error) {
	proxy := http.ProxyFromEnvironment
	if opt.Proxy != "" {
		proxyURL, err := parseProxy(opt.Proxy)
		if err != nil {
			return nil, err
		}
		proxy = http.ProxyURL(proxyURL)
	}
	transport := &http.Transport{
		Proxy:               proxy,
		DisableCompression:  true,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   15 * time.Minute,
		Jar:       jar,
	}

	ctx := opt.Context
	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for k, v := range config.FakeHeaders {
		req.Header.Set(k, v)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if _, ok := headers["Referer"]; !ok {
		req.Header.Set("Referer", url)
	}
	if opt.Cookie != "" {
		// parse cookies in Netscape HTTP cookie format
		cookies, _ := cookiemonster.ParseString(opt.Cookie)
		if len(cookies) > 0 {
			for _, c := range cookies {
				req.AddCookie(c)
			}
		} else {
			// cookie is not Netscape HTTP format, set it directly
			// a=b; c=d
			req.Header.Set("Cookie", opt.Cookie)
		}
	}

	if opt.UserAgent != "" {
		req.Header.Set("User-Agent", opt.UserAgent)
	}

	if opt.Refer != "" {
		req.Header.Set("Referer", opt.Refer)
	}

	var (
		res          *http.Response
		requestError error
	)
	for i := 0; ; i++ {
		res, requestError = client.Do(req)
		if requestError == nil && res.StatusCode < 400 {
			break
		} else if i+1 >= opt.RetryTimes {
			var err error
			if requestError != nil {
				err = errors.Wrap(requestError, "request error")
			} else {
				err = errors.Errorf("%s request error: HTTP %d", url, res.StatusCode)
			}
			return nil, errors.WithStack(err)
		}
		time.Sleep(1 * time.Second)
	}
	if opt.Debug {
		blue := color.New(color.FgBlue)
		fmt.Println()
		blue.Printf("URL:         ") // nolint
		fmt.Printf("%s\n", url)
		blue.Printf("Method:      ") // nolint
		fmt.Printf("%s\n", method)
		blue.Printf("Headers:     ")        // nolint
		pretty.Printf("%# v\n", req.Header) // nolint
		blue.Printf("Status Code: ")        // nolint
		if res.StatusCode >= 400 {
			color.Red("%d", res.StatusCode)
		} else {
			color.Green("%d", res.StatusCode)
		}
	}
	return res, nil
}

func parseProxy(raw string) (*url.URL, error) {
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}
	proxyURL, err := url.Parse(raw)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return proxyURL, nil
}

// Get get request
func Get(url, refer string, headers map[string]string) (string, error) {
	optionsMu.Lock()
	opt := currentOptionsUnlocked()
	optionsMu.Unlock()
	return GetWithOptions(url, refer, headers, opt)
}

// GetWithOptions runs a get request with explicit request options.
func GetWithOptions(url, refer string, headers map[string]string, opt Options) (string, error) {
	body, err := GetByteWithOptions(url, refer, headers, opt)
	return string(body), err
}

// GetByte get request
func GetByte(url, refer string, headers map[string]string) ([]byte, error) {
	optionsMu.Lock()
	opt := currentOptionsUnlocked()
	optionsMu.Unlock()
	return GetByteWithOptions(url, refer, headers, opt)
}

// GetByteWithOptions runs a get request with explicit request options.
func GetByteWithOptions(url, refer string, headers map[string]string, opt Options) ([]byte, error) {
	if headers == nil {
		headers = map[string]string{}
	}
	if refer != "" {
		headers["Referer"] = refer
	}
	res, err := RequestWithOptions(http.MethodGet, url, nil, headers, opt)
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

	body, err := io.ReadAll(reader)
	if err != nil && err != io.EOF {
		return nil, errors.WithStack(err)
	}
	return body, nil
}

// Headers return the HTTP Headers of the url
func Headers(url, refer string) (http.Header, error) {
	optionsMu.Lock()
	opt := currentOptionsUnlocked()
	optionsMu.Unlock()
	return HeadersWithOptions(url, refer, opt)
}

// HeadersWithOptions returns the HTTP headers with explicit request options.
func HeadersWithOptions(url, refer string, opt Options) (http.Header, error) {
	headers := map[string]string{
		"Referer": refer,
	}
	res, err := RequestWithOptions(http.MethodGet, url, nil, headers, opt)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer res.Body.Close() // nolint
	return res.Header, nil
}

// Size get size of the url
func Size(url, refer string) (int64, error) {
	optionsMu.Lock()
	opt := currentOptionsUnlocked()
	optionsMu.Unlock()
	return SizeWithOptions(url, refer, opt)
}

// SizeWithOptions gets size with explicit request options.
func SizeWithOptions(url, refer string, opt Options) (int64, error) {
	h, err := HeadersWithOptions(url, refer, opt)
	if err != nil {
		return 0, err
	}
	s := h.Get("Content-Length")
	if s == "" {
		return 0, errors.New("Content-Length is not present")
	}
	size, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return size, nil
}

// ContentType get Content-Type of the url
func ContentType(url, refer string) (string, error) {
	optionsMu.Lock()
	opt := currentOptionsUnlocked()
	optionsMu.Unlock()
	return ContentTypeWithOptions(url, refer, opt)
}

// ContentTypeWithOptions gets Content-Type with explicit request options.
func ContentTypeWithOptions(url, refer string, opt Options) (string, error) {
	h, err := HeadersWithOptions(url, refer, opt)
	if err != nil {
		return "", err
	}
	s := h.Get("Content-Type")
	// handle Content-Type like this: "text/html; charset=utf-8"
	return strings.Split(s, ";")[0], nil
}
