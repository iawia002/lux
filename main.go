package main

import (
	"bufio"
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
	flag.BoolVar(&config.ExtractedData, "j", false, "Print extracted data")
	flag.IntVar(&config.ThreadNumber, "n", 10, "The number of download thread")
	flag.StringVar(&config.File, "F", "", "URLs file")
	flag.IntVar(&config.PlaylistStart, "start", 1, "Playlist video to start at")
	flag.IntVar(&config.PlaylistEnd, "end", 0, "Playlist video to end at")
	flag.StringVar(
		&config.PlaylistItems, "items", "",
		"Playlist video items to download. Separated by commas like: 1,5,6",
	)
	flag.BoolVar(&config.Caption, "C", false, "Download captions")
	flag.StringVar(&config.Ccode, "ccode", "0590", "Youku ccode")
}

func download(videoURL string) {
	var domain string
	bilibiliShortLink := utils.MatchOneOf(videoURL, `^(av|ep)\d+`)
	if bilibiliShortLink != nil {
		bilibiliURL := map[string]string{
			"av": "https://www.bilibili.com/video/",
			"ep": "https://www.bilibili.com/bangumi/play/",
		}
		domain = "bilibili"
		videoURL = bilibiliURL[bilibiliShortLink[1]] + videoURL
	} else {
		u, err := url.ParseRequestURI(videoURL)
		if err != nil {
			log.Fatal(err)
		}
		domain = utils.Domain(u.Host)
	}
	switch domain {
	case "douyin", "iesdouyin":
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
	case "miaopai":
		extractors.Miaopai(videoURL)
	case "weibo":
		extractors.Weibo(videoURL)
	case "instagram":
		extractors.Instagram(videoURL)
	case "twitter":
		extractors.Twitter(videoURL)
	case "qq":
		extractors.QQ(videoURL)
	default:
		extractors.Universal(videoURL)
	}
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
	if config.File != "" {
		file, err := os.Open(config.File)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if scanner.Text() == "" {
				continue
			}
			args = append(args, scanner.Text())
		}
	}
	if len(args) < 1 {
		fmt.Println("Too few arguments")
		fmt.Println("Usage: annie [args] URLs...")
		flag.PrintDefaults()
		return
	}
	for _, videoURL := range args {
		download(videoURL)
	}
}
