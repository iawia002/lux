package main

import (
	"flag"
	"fmt"
	"net/url"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/extractors"
)

var (
	debug   bool
	version bool
)

func init() {
	flag.BoolVar(&debug, "d", false, "Debug mode")
	flag.BoolVar(&version, "v", false, "Show version")
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("error")
		return
	}
	videoUrl := args[0]
	u, err := url.ParseRequestURI(videoUrl)
	if err != nil {
		fmt.Println(err)
		return
	}
	domainPattern := `([a-z0-9][-a-z0-9]{0,62})\.` +
		`(com\.cn|com\.hk|` +
		`cn|com|net|edu|gov|biz|org|info|pro|name|xxx|xyz|` +
		`me|top|cc|tv|tt)`
	domain := downloader.Match1(domainPattern, u.Host)[1]
	switch domain {
	case "douyin":
		extractors.Douyin(videoUrl)
	default:
		fmt.Println("unsupported URL")
	}
}
