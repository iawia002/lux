package request

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// Request base request
func Request(method, url string, body io.Reader) *http.Response {
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
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Charset", "UTF-8,*;q=0.5")
	req.Header.Set("Accept-Encoding", "")
	req.Header.Set("Accept-Language", "en-US,en;q=0.8")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64; rv:51.0) Gecko/20100101 Firefox/51.0")
	req.Header.Set("Referer", url)
	res, err := client.Do(req)
	if err != nil {
		log.Print(url)
		log.Fatal(err)
	}
	return res
}

// Get get request
func Get(url string) string {
	res := Request("GET", url, nil)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	return string(body)
}
