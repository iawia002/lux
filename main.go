package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/extractors"
	"github.com/iawia002/annie/extractors/bilibili"
	"github.com/iawia002/annie/extractors/youtube"
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
	flag.StringVar(&config.Format, "f", "", "Select specific format to download")
	flag.StringVar(&config.OutputPath, "o", "", "Specify the output path")
	flag.StringVar(&config.OutputName, "O", "", "Specify the output file name")
}

func main() {
	flag.Parse()
	args := flag.Args()
	if config.Version {
		utils.PrintVersion()
		return
	}
	if config.Debug {
		utils.PrintVersion()
	}
	if len(args) < 1 {
		fmt.Printf("Too few arguments \n")
		fmt.Printf("Usage of %s: \n", os.Args[0])
		flag.PrintDefaults()
		return
	}
	videoURL := args[0]
	u, err := url.ParseRequestURI(videoURL)
	if err != nil {
		log.Fatal(err)
	}

	domain := utils.Domain(u.Host)
	switch domain {
	case "douyin":
		extractors.Douyin(videoURL)
	case "bilibili":
		bilibili.Download(videoURL)
	case "bcy":
		extractors.Bcy(videoURL)
	case "pixivision":
		extractors.Pixivision(videoURL)
	case "youku":
		extractors.Youku(videoURL)
	case "youtube", "youtu": // youtu.be
		youtube.Download(videoURL)
	case "iqiyi":
		extractors.Iqiyi(videoURL)
	case "mgtv":
		extractors.Mgtv(videoURL)
	case "tumblr":
		extractors.Tumblr(videoURL)
	case "vimeo":
		extractors.Vimeo(videoURL)
	case "facebook":
		extractors.Facebook(videoURL)
	case "douyu":
		extractors.Douyu(videoURL)
	default:
		extractors.Universal(videoURL)
	}
}
