package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/fatih/color"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/extractors/bcy"
	"github.com/iawia002/annie/extractors/bilibili"
	"github.com/iawia002/annie/extractors/douyin"
	"github.com/iawia002/annie/extractors/douyu"
	"github.com/iawia002/annie/extractors/facebook"
	"github.com/iawia002/annie/extractors/instagram"
	"github.com/iawia002/annie/extractors/iqiyi"
	"github.com/iawia002/annie/extractors/mgtv"
	"github.com/iawia002/annie/extractors/miaopai"
	"github.com/iawia002/annie/extractors/netease"
	"github.com/iawia002/annie/extractors/pixivision"
	"github.com/iawia002/annie/extractors/qq"
	"github.com/iawia002/annie/extractors/tumblr"
	"github.com/iawia002/annie/extractors/twitter"
	"github.com/iawia002/annie/extractors/universal"
	"github.com/iawia002/annie/extractors/vimeo"
	"github.com/iawia002/annie/extractors/weibo"
	"github.com/iawia002/annie/extractors/yinyuetai"
	"github.com/iawia002/annie/extractors/youku"
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
	flag.StringVar(&config.Stream, "f", "", "Select specific stream to download")
	flag.StringVar(&config.OutputPath, "o", "", "Specify the output path")
	flag.StringVar(&config.OutputName, "O", "", "Specify the output file name")
	flag.BoolVar(&config.ExtractedData, "j", false, "Print extracted data")
	flag.IntVar(&config.ChunkSizeMB, "cs", 0, "HTTP chunk size for downloading (in MB)")
	flag.BoolVar(&config.UseAria2RPC, "aria2", false, "Use Aria2 RPC to download")
	flag.StringVar(&config.Aria2Token, "aria2token", "", "Aria2 RPC Token")
	flag.StringVar(&config.Aria2Addr, "aria2addr", "localhost:6800", "Aria2 Address")
	flag.StringVar(&config.Aria2Method, "aria2method", "http", "Aria2 Method")
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
	flag.IntVar(
		&config.RetryTimes, "retry", 10, "How many times to retry when the download failed",
	)
	// youku
	flag.StringVar(&config.YoukuCcode, "ccode", "0590", "Youku ccode")
	flag.StringVar(
		&config.YoukuCkey,
		"ckey",
		"7B19C0AB12633B22E7FE81271162026020570708D6CC189E4924503C49D243A0DE6CD84A766832C2C99898FC5ED31F3709BB3CDD82C96492E721BDD381735026",
		"Youku ckey",
	)
	flag.StringVar(&config.YoukuPassword, "password", "", "Youku password")
	// youtube
	flag.BoolVar(&config.YouTubeStream2, "ytb-stream2", false, "Use data in url_encoded_fmt_stream_map")
}

func printError(url string, err error) {
	fmt.Printf(
		"Downloading %s error:\n%s\n",
		color.CyanString("%s", url), color.RedString("%v", err),
	)
}

func download(videoURL string) {
	var (
		domain string
		err    error
		data   []downloader.Data
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
			printError(videoURL, err)
			return
		}
		domain = utils.Domain(u.Host)
	}
	switch domain {
	case "douyin", "iesdouyin":
		data, err = douyin.Extract(videoURL)
	case "bilibili":
		data, err = bilibili.Extract(videoURL)
	case "bcy":
		data, err = bcy.Extract(videoURL)
	case "pixivision":
		data, err = pixivision.Extract(videoURL)
	case "youku":
		data, err = youku.Extract(videoURL)
	case "youtube", "youtu": // youtu.be
		data, err = youtube.Extract(videoURL)
	case "iqiyi":
		data, err = iqiyi.Extract(videoURL)
	case "mgtv":
		data, err = mgtv.Extract(videoURL)
	case "tumblr":
		data, err = tumblr.Extract(videoURL)
	case "vimeo":
		data, err = vimeo.Extract(videoURL)
	case "facebook":
		data, err = facebook.Extract(videoURL)
	case "douyu":
		data, err = douyu.Extract(videoURL)
	case "miaopai":
		data, err = miaopai.Extract(videoURL)
	case "163":
		data, err = netease.Extract(videoURL)
	case "weibo":
		data, err = weibo.Extract(videoURL)
	case "instagram":
		data, err = instagram.Extract(videoURL)
	case "twitter":
		data, err = twitter.Extract(videoURL)
	case "qq":
		data, err = qq.Extract(videoURL)
	case "yinyuetai":
		data, err = yinyuetai.Extract(videoURL)
	default:
		data, err = universal.Extract(videoURL)
	}
	if err != nil {
		// if this error occurs, it means that an error occurred before actually starting to extract data
		// (there is an error in the preparation step), and the data list is empty.
		printError(videoURL, err)
	}
	for _, item := range data {
		if item.Err != nil {
			// if this error occurs, the preparation step is normal, but the data extraction is wrong.
			// the data is an empty struct.
			printError(item.URL, item.Err)
			continue
		}
		err = downloader.Download(item, videoURL, config.ChunkSizeMB)
		if err != nil {
			printError(item.URL, err)
		}
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
		// read URL list from file
		file, err := os.Open(config.File)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			universalURL := strings.TrimSpace(scanner.Text())
			if universalURL == "" {
				continue
			}
			args = append(args, universalURL)
		}
	}
	if len(args) < 1 {
		fmt.Println("Too few arguments")
		fmt.Println("Usage: annie [args] URLs...")
		flag.PrintDefaults()
		return
	}
	if config.Cookie != "" {
		// If config.Cookie is a file path, convert it to a string to ensure
		// config.Cookie is always string
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
		download(strings.TrimSpace(videoURL))
	}
}
