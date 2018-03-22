package main

import (
	"flag"
	"fmt"
	"net/url"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/extractors"
	"github.com/iawia002/annie/utils"
)

func init() {
	flag.BoolVar(&config.Debug, "d", false, "Debug mode")
	flag.BoolVar(&config.Version, "v", false, "Show version")
	flag.BoolVar(&config.InfoOnly, "i", false, "Information only")
	flag.StringVar(&config.Cookie, "c", "", "Cookie")
	flag.BoolVar(&config.Playlist, "p", false, "Download playlist")
	flag.StringVar(&config.Refer, "r", "", "Use specified Referrer")
	flag.StringVar(&config.Proxy, "x", "", "HTTP proxy")
	flag.StringVar(&config.Socks5Proxy, "s", "", "SOCKS5 proxy")
}

func main() {
	flag.Parse()
	args := flag.Args()
	if config.Version {
		fmt.Printf(
			"annie: version %s, A simple and clean video downloader.\n", config.VERSION,
		)
		return
	}
	if len(args) < 1 {
		fmt.Println("error")
		return
	}
	videoURL := args[0]
	u, err := url.ParseRequestURI(videoURL)
	if err != nil {
		fmt.Println(err)
		return
	}

	domain := utils.Domain(u.Host)
	switch domain {
	case "douyin":
		extractors.Douyin(videoURL)
	case "bilibili":
		extractors.Bilibili(videoURL)
	case "bcy":
		extractors.Bcy(videoURL)
	case "pixivision":
		extractors.Pixivision(videoURL)
	case "youku":
		extractors.Youku(videoURL)
	case "youtube", "youtu": // youtu.be
		extractors.Youtube(videoURL)
	case "iqiyi":
		extractors.Iqiyi(videoURL)
	case "mgtv":
		extractors.Mgtv(videoURL)
	case "tumblr":
		extractors.Tumblr(videoURL)
	case "vimeo":
		extractors.Vimeo(videoURL)
	default:
		extractors.Universal(videoURL)
	}
}
