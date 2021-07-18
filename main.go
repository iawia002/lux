package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/go-rod/rod"
	"github.com/urfave/cli/v2"

	"github.com/iawia002/annie/cookier"
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/extractors"
	"github.com/iawia002/annie/extractors/types"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		blue := color.New(color.FgBlue)
		cyan := color.New(color.FgCyan)
		fmt.Fprintf(
			color.Output,
			"\n%s: version %s, A fast, simple and clean video downloader.\n\n",
			cyan.Sprintf("annie"),
			blue.Sprintf(c.App.Version),
		)
	}

	app := &cli.App{
		Name:    "annie",
		Usage:   "A fast, simple and clean video downloader.",
		Version: "v0.11.0",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Usage:   "Debug mode",
			},
			&cli.BoolFlag{
				Name:    "info",
				Aliases: []string{"i"},
				Usage:   "Information only",
			},
			&cli.BoolFlag{
				Name:    "json",
				Aliases: []string{"j"},
				Usage:   "Print extracted JSON data",
			},

			&cli.StringFlag{
				Name:    "cookie",
				Aliases: []string{"c"},
				Usage:   "Cookie",
			},
			&cli.BoolFlag{
				Name:    "playlist",
				Aliases: []string{"p"},
				Usage:   "Download playlist",
			},
			&cli.StringFlag{
				Name:    "refer",
				Aliases: []string{"r"},
				Usage:   "Use specified Referrer",
			},
			&cli.StringFlag{
				Name:    "stream-format",
				Aliases: []string{"f"},
				Usage:   "Select specific stream to download",
			},
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"F"},
				Usage:   "URLs file path",
			},
			&cli.StringFlag{
				Name:    "output-path",
				Aliases: []string{"o"},
				Usage:   "Specify the output path",
			},
			&cli.StringFlag{
				Name:    "output-name",
				Aliases: []string{"O"},
				Usage:   "Specify the output file name",
			},
			&cli.UintFlag{
				Name:  "file-name-length",
				Value: 255,
				Usage: "The maximum length of a file name, 0 means unlimited",
			},
			&cli.BoolFlag{
				Name:    "caption",
				Aliases: []string{"C"},
				Usage:   "Download captions",
			},

			&cli.UintFlag{
				Name:  "start",
				Value: 1,
				Usage: "Define the starting item of a playlist or a file input",
			},
			&cli.UintFlag{
				Name:  "end",
				Value: 0,
				Usage: "Define the ending item of a playlist or a file input",
			},
			&cli.StringFlag{
				Name:  "items",
				Usage: "Define wanted items from a file or playlist. Separated by commas like: 1,5,6,8-10",
			},

			&cli.BoolFlag{
				Name:    "multi-thread",
				Aliases: []string{"m"},
				Usage:   "Multiple threads to download single video",
			},
			&cli.UintFlag{
				Name:  "retry",
				Value: 10,
				Usage: "How many times to retry when the download failed",
			},
			&cli.UintFlag{
				Name:    "chunk-size",
				Aliases: []string{"cs"},
				Value:   1,
				Usage:   "HTTP chunk size for downloading (in MB)",
			},
			&cli.UintFlag{
				Name:    "thread",
				Aliases: []string{"n"},
				Value:   10,
				Usage:   "The number of download thread (only works for multiple-parts video)",
			},

			// Aria2
			&cli.BoolFlag{
				Name:  "aria2",
				Usage: "Use Aria2 RPC to download",
			},
			&cli.StringFlag{
				Name:  "aria2-token",
				Usage: "Aria2 RPC Token",
			},
			&cli.StringFlag{
				Name:  "aria2-addr",
				Value: "localhost:6800",
				Usage: "Aria2 Address",
			},
			&cli.StringFlag{
				Name:  "aria2-method",
				Value: "http",
				Usage: "Aria2 Method",
			},

			// youku
			&cli.StringFlag{
				Name:    "youku-ccode",
				Aliases: []string{"ccode"},
				Value:   "0532",
				Usage:   "Youku ccode",
			},
			&cli.StringFlag{
				Name:    "youku-ckey",
				Aliases: []string{"ckey"},
				Value:   "7B19C0AB12633B22E7FE81271162026020570708D6CC189E4924503C49D243A0DE6CD84A766832C2C99898FC5ED31F3709BB3CDD82C96492E721BDD381735026",
				Usage:   "Youku ckey",
			},
			&cli.StringFlag{
				Name:    "youku-password",
				Aliases: []string{"password"},
				Usage:   "Youku password",
			},

			&cli.BoolFlag{
				Name:    "episode-title-only",
				Aliases: []string{"eto"},
				Usage:   "File name of each bilibili episode doesn't include the playlist title",
			},
		},
		Action: func(c *cli.Context) error {
			args := c.Args().Slice()

			if c.Bool("debug") {
				cli.VersionPrinter(c)
			}

			if file := c.String("file"); file != "" {
				f, err := os.Open(file)
				if err != nil {
					return err
				}
				defer f.Close() // nolint

				fileItems := utils.ParseInputFile(f, c.String("items"), int(c.Uint("start")), int(c.Uint("end")))
				args = append(args, fileItems...)
			}

			if len(args) < 1 {
				return errors.New("too few arguments")
			}

			cookie := c.String("cookie")
			if cookie != "" {
				// If cookie is a file path, convert it to a string to ensure cookie is always string
				if _, fileErr := os.Stat(cookie); fileErr == nil {
					// Cookie is a file
					data, err := ioutil.ReadFile(cookie)
					if err != nil {
						return err
					}
					cookie = strings.TrimSpace(string(data))
				}
			} else {
				// Try to use current user's cookie if possible, if failed empty cookie will be used
				_ = rod.Try(func() {
					cookie = cookier.Get(args...)
				})
			}

			request.SetOptions(request.Options{
				RetryTimes: int(c.Uint("retry")),
				Cookie:     cookie,
				Refer:      c.String("refer"),
				Debug:      c.Bool("debug"),
			})

			var isErr bool
			for _, videoURL := range args {
				if err := download(c, videoURL); err != nil {
					fmt.Fprintf(
						color.Output,
						"Downloading %s error:\n%s\n",
						color.CyanString("%s", videoURL), color.RedString("%v", err),
					)
					isErr = true
				}
			}
			if isErr {
				return cli.Exit("", 1)
			}
			return nil
		},
		EnableBashCompletion: true,
	}

	sort.Sort(cli.FlagsByName(app.Flags))

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(
			color.Output,
			"Run %s failed: %s\n",
			color.CyanString("%s", app.Name), color.RedString("%v", err),
		)
		os.Exit(1)
	}
}

func download(c *cli.Context, videoURL string) error {
	data, err := extractors.Extract(videoURL, types.Options{
		Playlist:         c.Bool("playlist"),
		Items:            c.String("items"),
		ItemStart:        int(c.Uint("start")),
		ItemEnd:          int(c.Uint("end")),
		ThreadNumber:     int(c.Uint("thread")),
		EpisodeTitleOnly: c.Bool("episode-title-only"),
		Cookie:           c.String("cookie"),
		YoukuCcode:       c.String("youku-ccode"),
		YoukuCkey:        c.String("youku-ckey"),
		YoukuPassword:    c.String("youku-password"),
	})
	if err != nil {
		// if this error occurs, it means that an error occurred before actually starting to extract data
		// (there is an error in the preparation step), and the data list is empty.
		return err
	}

	if c.Bool("json") {
		jsonData, err := json.MarshalIndent(data, "", "\t")
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", jsonData)
		return nil
	}

	defaultDownloader := downloader.New(downloader.Options{
		InfoOnly:       c.Bool("info"),
		Stream:         c.String("stream-format"),
		Refer:          c.String("refer"),
		OutputPath:     c.String("output-path"),
		OutputName:     c.String("output-name"),
		FileNameLength: int(c.Uint("file-name-length")),
		Caption:        c.Bool("caption"),
		MultiThread:    c.Bool("multi-thread"),
		ThreadNumber:   int(c.Uint("thread")),
		RetryTimes:     int(c.Uint("retry")),
		ChunkSizeMB:    int(c.Uint("chunk-size")),
		UseAria2RPC:    c.Bool("aria2"),
		Aria2Token:     c.String("aria2-token"),
		Aria2Method:    c.String("aria2-method"),
		Aria2Addr:      c.String("aria2-addr"),
	})
	errors := make([]error, 0)
	for _, item := range data {
		if item.Err != nil {
			// if this error occurs, the preparation step is normal, but the data extraction is wrong.
			// the data is an empty struct.
			errors = append(errors, item.Err)
			continue
		}
		if err = defaultDownloader.Download(item); err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) != 0 {
		return errors[0]
	}
	return nil
}
