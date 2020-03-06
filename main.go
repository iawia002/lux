package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/rs/zerolog"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/extractors/bcy"
	"github.com/iawia002/annie/extractors/bilibili"
	"github.com/iawia002/annie/extractors/douyin"
	"github.com/iawia002/annie/extractors/douyu"
	"github.com/iawia002/annie/extractors/facebook"
	"github.com/iawia002/annie/extractors/geekbang"
	"github.com/iawia002/annie/extractors/instagram"
	"github.com/iawia002/annie/extractors/iqiyi"
	"github.com/iawia002/annie/extractors/mgtv"
	"github.com/iawia002/annie/extractors/miaopai"
	"github.com/iawia002/annie/extractors/netease"
	"github.com/iawia002/annie/extractors/pixivision"
	"github.com/iawia002/annie/extractors/pornhub"
	"github.com/iawia002/annie/extractors/qq"
	"github.com/iawia002/annie/extractors/tangdou"
	"github.com/iawia002/annie/extractors/tiktok"
	"github.com/iawia002/annie/extractors/tumblr"
	"github.com/iawia002/annie/extractors/twitter"
	"github.com/iawia002/annie/extractors/udn"
	"github.com/iawia002/annie/extractors/universal"
	"github.com/iawia002/annie/extractors/vimeo"
	"github.com/iawia002/annie/extractors/weibo"
	"github.com/iawia002/annie/extractors/xvideos"
	"github.com/iawia002/annie/extractors/yinyuetai"
	"github.com/iawia002/annie/extractors/youku"
	"github.com/iawia002/annie/extractors/youtube"
	"github.com/iawia002/annie/utils"
)

func init() {
	flag.BoolVar(&config.MultiThread, "m", false, "Multiple threads to download single video")
	flag.BoolVar(&config.Debug, "d", false, "Debug mode")
	flag.BoolVar(&config.Version, "v", false, "Show version")
	flag.BoolVar(&config.InfoOnly, "i", false, "Information only")
	flag.StringVar(&config.Cookie, "c", "", "Cookie")
	flag.BoolVar(&config.Playlist, "p", false, "Download playlist")
	flag.StringVar(&config.Refer, "r", "", "Use specified Referrer")
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
	flag.IntVar(&config.ItemStart, "start", 1, "Define the starting item of a playlist or a file input")
	flag.IntVar(&config.ItemEnd, "end", 0, "Define the ending item of a playlist or a file input")
	flag.StringVar(
		&config.Items, "items", "",
		"Define wanted items from a file or playlist. Separated by commas like: 1,5,6,8-10",
	)
	flag.BoolVar(&config.EpisodeTitleOnly, "eto", false, "File name of each bilibili episode doesn't include the playlist title")
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
}

func download(videoURL string) error {
	var (
		domain string
		err    error
		data   []downloader.Data
	)
	bilibiliShortLink := utils.MatchOneOf(videoURL, `^(av|ep)\d+`)
	if bilibiliShortLink != nil && len(bilibiliShortLink) > 1 {
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
	case "tangdou":
		data, err = tangdou.Extract(videoURL)
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
	case "geekbang":
		data, err = geekbang.Extract(videoURL)
	case "pornhub":
		data, err = pornhub.Extract(videoURL)
	case "xvideos":
		data, err = xvideos.Extract(videoURL)
	case "udn":
		data, err = udn.Extract(videoURL)
	case "tiktok":
		data, err = tiktok.Extract(videoURL)
	default:
		data, err = universal.Extract(videoURL)
	}
	if err != nil {
		// if this error occurs, it means that an error occurred before actually starting to extract data
		// (there is an error in the preparation step), and the data list is empty.
		return err
	}

	if config.ExtractedData {
		jsonData, err := json.MarshalIndent(data, "", "\t")
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", jsonData)
		return nil
	}

	errors := make([]error, 0)
	for _, item := range data {
		if item.Err != nil {
			// if this error occurs, the preparation step is normal, but the data extraction is wrong.
			// the data is an empty struct.
			errors = append(errors, item.Err)
			continue
		}
		err = downloader.Download(item, videoURL, config.ChunkSizeMB)
		if err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) != 0 {
		return errors[0]
	}
	return nil
}

func printError(url string, err error) {
	fmt.Printf(
		"Downloading %s error:\n%s\n",
		color.CyanString("%s", url), color.RedString("%v", err),
	)
}

func main() {
	flag.Parse()
	args := flag.Args()
	if config.Version {
		utils.PrintVersion()
		return
	}
	// introduced by "github.com/rylio/ytdl"
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if config.Debug {
		utils.PrintVersion()
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	if config.File != "" {
		file, err := os.Open(config.File)
		if err != nil {
			fmt.Printf("Error %v", err)
			return
		}
		defer file.Close()

		fileItems := utils.ParseInputFile(file)
		args = append(args, fileItems...)
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
			config.Cookie = strings.TrimSpace(string(data))
		}
	}
	var isErr bool
	for _, videoURL := range args {
		u := strings.TrimSpace(videoURL)
		if err := download(u); err != nil {
			printError(u, err)
			isErr = true
		}
	}
	if isErr {
		os.Exit(1)
	}
}
