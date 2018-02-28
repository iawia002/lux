package request

import (
	"compress/gzip"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/iawia002/annie/config"
)

// Request base request
func Request(
	method, url string, body io.Reader, headers map[string]string,
) *http.Response {
	client := &http.Client{
		Timeout: time.Second * 100,
		Transport: &http.Transport{
			DisableCompression: true,
		},
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Print(url)
		log.Fatal(err)
	}
	for k, v := range config.FAKE_HEADERS {
		req.Header.Set(k, v)
	}
	req.Header.Set("Referer", url)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Print(url)
		log.Fatal(err)
	}
	return res
}

// Get get request
func Get(url string) string {
	res := Request("GET", url, nil, nil)
	defer res.Body.Close()
	var reader io.ReadCloser
	if res.Header.Get("Content-Encoding") == "gzip" {
		reader, _ = gzip.NewReader(res.Body)
	} else {
		reader = res.Body
	}
	body, _ := ioutil.ReadAll(reader)
	return string(body)
}

// Size get size of the url
func Size(url, refer string) int64 {
	headers := map[string]string{
		"Referer": refer,
	}
	res := Request("GET", url, nil, headers)
	defer res.Body.Close()
	s := res.Header.Get("Content-Length")
	size, _ := strconv.ParseInt(s, 10, 64)
	return size
}
