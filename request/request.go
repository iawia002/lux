package request

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
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
	body, _ := ioutil.ReadAll(res.Body)
	return string(body)
}
