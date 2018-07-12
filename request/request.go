package request

import (
	"compress/flate"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	netURL "net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/kr/pretty"
	"golang.org/x/net/proxy"

	"github.com/iawia002/annie/config"
)

// Request base request
func Request(
	method, url string, body io.Reader, headers map[string]string,
) *http.Response {
	transport := &http.Transport{
		DisableCompression:  true,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}
	if config.Proxy != "" {
		var httpProxy, err = netURL.Parse(config.Proxy)
		if err != nil {
			panic(err)
		}
		transport.Proxy = http.ProxyURL(httpProxy)
	}
	if config.Socks5Proxy != "" {
		dialer, err := proxy.SOCKS5(
			"tcp",
			config.Socks5Proxy,
			nil,
			&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			},
		)
		if err != nil {
			panic(err)
		}
		transport.Dial = dialer.Dial
	}
	client := &http.Client{
		Transport: transport,
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Print(url)
		panic(err)
	}
	for k, v := range config.FakeHeaders {
		req.Header.Set(k, v)
	}
	req.Header.Set("Referer", url)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if config.Cookie != "" {
		var cookie string
		if _, fileErr := os.Stat(config.Cookie); fileErr == nil {
			// Cookie is a file
			data, _ := ioutil.ReadFile(config.Cookie)
			cookie = string(data)
		} else {
			// Just strings
			cookie = config.Cookie
		}
		req.Header.Set("Cookie", cookie)
	}
	if config.Refer != "" {
		req.Header.Set("Referer", config.Refer)
	}
	retryTimes := 3
	var (
		res          *http.Response
		requestError error
	)
	for i := 0; i < retryTimes; i++ {
		res, requestError = client.Do(req)
		if requestError == nil {
			break
		}
		if requestError != nil && i+1 == retryTimes {
			log.Print(url)
			panic(requestError)
		}
		time.Sleep(1 * time.Second)
	}
	if config.Debug {
		blue := color.New(color.FgBlue)
		fmt.Println()
		blue.Printf("URL:         ")
		fmt.Printf("%s\n", url)
		blue.Printf("Method:      ")
		fmt.Printf("%s\n", method)
		blue.Printf("Headers:     ")
		pretty.Printf("%# v\n", req.Header)
		blue.Printf("Status Code: ")
		if res.StatusCode >= 400 {
			color.Red("%d", res.StatusCode)
		} else {
			color.Green("%d", res.StatusCode)
		}
	}
	return res
}

// Get get request
func Get(url, refer string, headers map[string]string) string {
	if headers == nil {
		headers = map[string]string{}
	}
	if refer != "" {
		headers["Referer"] = refer
	}
	res := Request("GET", url, nil, headers)
	var reader io.ReadCloser
	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		reader, _ = gzip.NewReader(res.Body)
	case "deflate":
		reader = flate.NewReader(res.Body)
	default:
		reader = res.Body
	}
	defer res.Body.Close()
	defer reader.Close()
	body, _ := ioutil.ReadAll(reader)
	return string(body)
}

// Headers return the HTTP Headers of the url
func Headers(url, refer string) http.Header {
	headers := map[string]string{
		"Referer": refer,
	}
	res := Request("GET", url, nil, headers)
	return res.Header
}

// Size get size of the url
func Size(url, refer string) int64 {
	h := Headers(url, refer)
	s := h.Get("Content-Length")
	size, _ := strconv.ParseInt(s, 10, 64)
	return size
}

// ContentType get Content-Type of the url
func ContentType(url, refer string) string {
	h := Headers(url, refer)
	s := h.Get("Content-Type")
	// handle Content-Type like this: "text/html; charset=utf-8"
	return strings.Split(s, ";")[0]
}
