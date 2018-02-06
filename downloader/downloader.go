package downloader

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/cheggaaa/pb"

	"github.com/iawia002/annie/extractors/utils"
)

func Match1(pattern, text string) []string {
	re := regexp.MustCompile(pattern)
	value := re.FindStringSubmatch(text)
	return value
}

func request(method, url string, body io.Reader) *http.Response {
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

func Get(url string) string {
	res := request("GET", url, nil)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	return string(body)
}

func UrlSize(url string) int {
	res := request("GET", url, nil)
	defer res.Body.Close()
	s := res.Header.Get("Content-Length")
	size, _ := strconv.Atoi(s)
	return size
}

func PrintInfo(data utils.VideoData) {
	fmt.Println()
	fmt.Println(" Site:   ", data.Site)
	fmt.Println("Title:   ", data.Title)
	fmt.Println(" Type:   ", data.Ext)
	fmt.Printf(" Size:    %.2f MiB (%d Bytes)\n", float64(data.Size)/1000000.0, data.Size)
	fmt.Println()
}

func UlrSave(data utils.VideoData) {
	PrintInfo(data)
	res := request("GET", data.Url, nil)
	defer res.Body.Close()
	file, _ := os.Create(data.Title + "." + data.Ext)
	bar := pb.New(int(data.Size)).SetUnits(pb.U_BYTES).SetRefreshRate(time.Millisecond * 10)
	bar.Start()
	bar.ShowSpeed = true
	bar.ShowFinalTime = true
	bar.SetMaxWidth(1000)
	writer := io.MultiWriter(file, bar)
	io.Copy(writer, res.Body)
	bar.Finish()
}
