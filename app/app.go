package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"

	"github.com/iawia002/lux/downloader"
	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

const Name = "lux"

var version = "v0.0.0"

func init() {
	cli.VersionPrinter = func(c *cli.Context) {
		blue := color.New(color.FgBlue)
		cyan := color.New(color.FgCyan)
		fmt.Fprintf(
			color.Output,
			"\n%s: version %s, A fast and simple video downloader.\n\n",
			cyan.Sprintf(Name),
			blue.Sprintf(c.App.Version),
		)
	}
}

func initConfig(configPath string) error {
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		// Search config in current dir or $HOME/.config/lux/config.yaml
		home, _ := os.UserHomeDir()
		viper.AddConfigPath(".")
		viper.AddConfigPath(filepath.Join(home, ".config", "lux"))
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}
	return nil
}

// Priority: cmd args > config file > default value
func getString(c *cli.Context, name string) string {
	if c.IsSet(name) {
		return c.String(name)
	}
	return viper.GetString(name)
}

func getBool(c *cli.Context, name string) bool {
	if c.IsSet(name) {
		return c.Bool(name)
	}
	return viper.GetBool(name)
}

func getUint(c *cli.Context, name string) uint {
	if c.IsSet(name) {
		return c.Uint(name)
	}
	return viper.GetUint(name)
}

func New() *cli.App {
	app := &cli.App{
		Name:    Name,
		Usage:   "A fast and simple video downloader.",
		Version: version,
		Flags: []cli.Flag{
			// Use config file to run lux
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"cfg"},
				Usage:   "Path to config file",
			},
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Usage:   "Debug mode",
			},
			&cli.BoolFlag{
				Name:    "silent",
				Aliases: []string{"s"},
				Usage:   "Minimum outputs",
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
				Name:    "user-agent",
				Aliases: []string{"u"},
				Usage:   "Use specified User-Agent",
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
			&cli.BoolFlag{
				Name:    "audio-only",
				Aliases: []string{"ao"},
				Usage:   "Download audio only at best quality",
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
			&cli.BoolFlag{
				Name:    "embed-subtitle",
				Aliases: []string{"embed"},
				Usage:   "Embed subtitles into the video (requires ffmpeg)",
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
				Value:   "0502",
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
			// load config file
			if err := initConfig(c.String("config")); err != nil {
				return err
			}

			args := c.Args().Slice()

			if getBool(c, "debug") {
				cli.VersionPrinter(c)
			}

			if file := getString(c, "file"); file != "" {
				f, err := os.Open(file)
				if err != nil {
					return err
				}
				defer f.Close() // nolint

				fileItems := utils.ParseInputFile(f, getString(c, "items"), int(getUint(c, "start")), int(getUint(c, "end")))
				args = append(args, fileItems...)
			}

			if len(args) < 1 {
				return errors.New("too few arguments")
			}

			cookie := getString(c, "cookie")
			if cookie != "" {
				// If cookie is a file path, convert it to a string to ensure cookie is always string
				if _, fileErr := os.Stat(cookie); fileErr == nil {
					// Cookie is a file
					data, err := os.ReadFile(cookie)
					if err != nil {
						return err
					}
					cookie = strings.TrimSpace(string(data))
				}
			}

			request.SetOptions(request.Options{
				RetryTimes: int(getUint(c, "retry")),
				Cookie:     cookie,
				UserAgent:  getString(c, "user-agent"),
				Refer:      getString(c, "refer"),
				Debug:      getBool(c, "debug"),
				Silent:     getBool(c, "silent"),
			})

			var isErr bool
			for _, videoURL := range args {
				if err := download(c, videoURL); err != nil {
					fmt.Fprintf(
						color.Output,
						"Downloading %s error:\n",
						color.CyanString("%s", videoURL),
					)
					fmt.Printf("%+v\n", err)
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
	return app
}

func download(c *cli.Context, videoURL string) error {
	data, err := extractors.Extract(videoURL, extractors.Options{
		Playlist:         getBool(c, "playlist"),
		Items:            getString(c, "items"),
		ItemStart:        int(getUint(c, "start")),
		ItemEnd:          int(getUint(c, "end")),
		ThreadNumber:     int(getUint(c, "thread")),
		EpisodeTitleOnly: getBool(c, "episode-title-only"),
		Cookie:           getString(c, "cookie"),
		YoukuCcode:       getString(c, "youku-ccode"),
		YoukuCkey:        getString(c, "youku-ckey"),
		YoukuPassword:    getString(c, "youku-password"),
	})
	if err != nil {
		// if this error occurs, it means that an error occurred before actually starting to extract data
		// (there is an error in the preparation step), and the data list is empty.
		return err
	}

	if getBool(c, "json") {
		e := json.NewEncoder(os.Stdout)
		e.SetIndent("", "\t")
		e.SetEscapeHTML(false)
		if err := e.Encode(data); err != nil {
			return err
		}

		return nil
	}

	defaultDownloader := downloader.New(downloader.Options{
		Silent:         getBool(c, "silent"),
		InfoOnly:       getBool(c, "info"),
		Stream:         getString(c, "stream-format"),
		AudioOnly:      getBool(c, "audio-only"),
		Refer:          getString(c, "refer"),
		OutputPath:     getString(c, "output-path"),
		OutputName:     getString(c, "output-name"),
		FileNameLength: int(getUint(c, "file-name-length")),
		Caption:        getBool(c, "caption"),
		EmbedSubtitle:  getBool(c, "embed-subtitle"),
		MultiThread:    getBool(c, "multi-thread"),
		ThreadNumber:   int(getUint(c, "thread")),
		RetryTimes:     int(getUint(c, "retry")),
		ChunkSizeMB:    int(getUint(c, "chunk-size")),
		UseAria2RPC:    getBool(c, "aria2"),
		Aria2Token:     getString(c, "aria2-token"),
		Aria2Method:    getString(c, "aria2-method"),
		Aria2Addr:      getString(c, "aria2-addr"),
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
