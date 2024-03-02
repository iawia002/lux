package request

import (
	"compress/flate"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"time"

	cookiemonster "github.com/MercuryEngineering/CookieMonster"
	"github.com/fatih/color"
	"github.com/kr/pretty"
	"github.com/pkg/errors"

	"github.com/iawia002/lux/config"
)

var (
	retryTimes int
	rawCookie  string
	userAgent  string
	refer      string
	debug      bool
)

// Options defines common request options.
type Options struct {
	RetryTimes int
	Cookie     string
	UserAgent  string
	Refer      string
	Debug      bool
	Silent     bool
}

// SetOptions sets the common request option.
func SetOptions(opt Options) {
	retryTimes = opt.RetryTimes
	rawCookie = opt.Cookie
	userAgent = opt.UserAgent
	refer = opt.Refer
	debug = opt.Debug
}

// Request base request
func Request(method, url string, body io.Reader, headers map[string]string) (*http.Response, error) {
	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
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

	req, err := http.NewRequest(method, url, body)
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
	if rawCookie != "" {
		// parse cookies in Netscape HTTP cookie format
		cookies, _ := cookiemonster.ParseString(rawCookie)
		if len(cookies) > 0 {
			for _, c := range cookies {
				req.AddCookie(c)
			}
		} else {
			// cookie is not Netscape HTTP format, set it directly
			// a=b; c=d
			req.Header.Set("Cookie", rawCookie)
		}
	}

	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}

	if refer != "" {
		req.Header.Set("Referer", refer)
	}

	var (
		res          *http.Response
		requestError error
	)
	for i := 0; ; i++ {
		res, requestError = client.Do(req)
		if requestError == nil && res.StatusCode < 400 {
			break
		} else if i+1 >= retryTimes {
			var err error
			if requestError != nil {
				err = errors.Errorf("request error: %v", requestError)
			} else {
				err = errors.Errorf("%s request error: HTTP %d", url, res.StatusCode)
			}
			return nil, errors.WithStack(err)
		}
		time.Sleep(1 * time.Second)
	}
	if debug {
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

// Get get request
func Get(url, refer string, headers map[string]string) (string, error) {
	body, err := GetByte(url, refer, headers)
	return string(body), err
}

// GetByte get request
func GetByte(url, refer string, headers map[string]string) ([]byte, error) {
	if headers == nil {
		headers = map[string]string{}
	}
	if refer != "" {
		headers["Referer"] = refer
	}
	res, err := Request(http.MethodGet, url, nil, headers)
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
	headers := map[string]string{
		"Referer": refer,
	}
	res, err := Request(http.MethodGet, url, nil, headers)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer res.Body.Close() // nolint
	return res.Header, nil
}

// Size get size of the url
func Size(url, refer string) (int64, error) {
	h, err := Headers(url, refer)
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
	h, err := Headers(url, refer)
	if err != nil {
		return "", err
	}
	s := h.Get("Content-Type")
	// handle Content-Type like this: "text/html; charset=utf-8"
	return strings.Split(s, ";")[0], nil
}
