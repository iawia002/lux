package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/fatih/color"

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
	flag.IntVar(
		&config.ThreadNumber, "n", 10, "The number of download thread (only works for multiple-parts video)",
	)
	flag.StringVar(&config.File, "F", "", "URLs file path")
	flag.IntVar(&config.PlaylistStart, "start", 1, "Playlist video to start at")
	flag.IntVar(&config.PlaylistEnd, "end", 0, "Playlist video to end at")
	flag.StringVar(
		&config.PlaylistItems, "items", "",
		"Playlist video items to download. Separated by commas like: 1,5,6",
	)
	flag.BoolVar(&config.Caption, "C", false, "Download captions")
	flag.StringVar(&config.Ccode, "ccode", "0103010102", "Youku ccode")
	flag.IntVar(
		&config.RetryTimes, "retry", 100, "How many times to retry when the download failed",
	)
}

func download(videoURL string) error {
	var (
		domain string
		err    error
	)
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
			return err
		}
		domain = utils.Domain(u.Host)
	}
	switch domain {
	case "douyin", "iesdouyin":
		_, err = extractors.Douyin(videoURL)
	case "bilibili":
		err = bilibili.Download(videoURL)
	case "bcy":
		_, err = extractors.Bcy(videoURL)
	case "pixivision":
		_, err = extractors.Pixivision(videoURL)
	case "youku":
		_, err = extractors.Youku(videoURL)
	case "youtube", "youtu": // youtu.be
		err = youtube.Download(videoURL)
	case "iqiyi":
		_, err = extractors.Iqiyi(videoURL)
	case "mgtv":
		_, err = extractors.Mgtv(videoURL)
	case "tumblr":
		_, err = extractors.Tumblr(videoURL)
	case "vimeo":
		_, err = extractors.Vimeo(videoURL)
	case "facebook":
		_, err = extractors.Facebook(videoURL)
	case "douyu":
		_, err = extractors.Douyu(videoURL)
	case "miaopai":
		_, err = extractors.Miaopai(videoURL)
	case "weibo":
		_, err = extractors.Weibo(videoURL)
	case "instagram":
		_, err = extractors.Instagram(videoURL)
	case "twitter":
		_, err = extractors.Twitter(videoURL)
	case "qq":
		_, err = extractors.QQ(videoURL)
	default:
		_, err = extractors.Universal(videoURL)
	}
	return err
}

func main() {
	var err error
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
			fmt.Println(err)
			return
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
	if config.Cookie != "" {
		if _, fileErr := os.Stat(config.Cookie); fileErr == nil {
			// Cookie is a file
			data, err := ioutil.ReadFile(config.Cookie)
			if err != nil {
				color.Red("%v", err)
				return
			}
			config.Cookie = string(data)
		}
	}
	for _, videoURL := range args {
		err = download(videoURL)
		if err != nil {
			fmt.Printf(
				"Downloading %s error:\n%s\n",
				color.CyanString("%s", videoURL), color.RedString("%v", err),
			)
		}
	}
}
