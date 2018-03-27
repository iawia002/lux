package bilibili

import (
	"testing"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/test"
)

func TestBilibili(t *testing.T) {
	config.InfoOnly = true
	tests := []struct {
		name     string
		args     test.Args
		playlist bool
	}{
		{
			name: "normal test",
			args: test.Args{
				URL:     "https://www.bilibili.com/video/av20203945/",
				Bangumi: false,
				Title:   "【2018拜年祭单品】相遇day by day",
				Quality: "高清 1080P",
			},
		},
		{
			name: "normal test",
			args: test.Args{
				URL:     "https://www.bilibili.com/video/av20203945/",
				Bangumi: false,
				Title:   "【2018拜年祭单品】相遇day by day",
				Quality: "高清 1080P",
			},
			playlist: true,
		},
		{
			name: "bangumi test",
			args: test.Args{
				URL:     "https://www.bilibili.com/bangumi/play/ep167000",
				Bangumi: true,
				Title:   "狐妖小红娘：第70话 苏苏智商上线",
				Quality: "高清 1080P",
			},
		},
		{
			name: "bangumi playlist test",
			args: test.Args{
				URL:     "https://www.bilibili.com/bangumi/play/ss5050",
				Bangumi: true,
				Title:   "一人之下：第1话 异人刀兵起，道炁携阴阳",
				Quality: "高清 1080P",
			},
			playlist: true,
		},
		{
			name: "playlist test",
			args: test.Args{
				URL:     "https://www.bilibili.com/video/av16907446/",
				Bangumi: false,
				Title:   "\"不要相信歌词，他们为了押韵什么都干得出来\"",
				Quality: "高清 720P",
			},
			playlist: true,
		},
		{
			name: "bangumi test",
			args: test.Args{
				URL:     "https://www.bilibili.com/bangumi/play/ss12044",
				Bangumi: true,
				Title:   "你的名字。",
			},
		},
		// {
		// 	name: "playlist test",
		// 	args: test.Args{
		// 		URL:     "https://www.bilibili.com/video/av20827366/",
		// 		Title:   "【极限2K画质 60fps】这可能是我做过最美的miku了【boomclap布料解算版】 1080P版",
		// 		Size:    38664181,
		// 		Quality: "高清 1080P",
		// 	},
		// },
		// {
		// 	name: "playlist test",
		// 	args: test.Args{
		// 		URL:     "https://www.bilibili.com/video/av20827366/?p=1",
		// 		Title:   "【极限2K画质 60fps】这可能是我做过最美的miku了【boomclap布料解算版】 1080P版",
		// 		Size:    38664181,
		// 		Quality: "高清 1080P",
		// 	},
		// },
		// {
		// 	name: "playlist test",
		// 	args: test.Args{
		// 		URL:     "https://www.bilibili.com/video/av20827366/?p=2",
		// 		Title:   "极限2K画质 60fps】这可能是我做过最美的miku了【boomclap布料解算版】 2K版",
		// 		Size:    68503929,
		// 		Quality: "高清 1080P",
		// 	},
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data downloader.VideoData
			options := bilibiliOptions{
				Bangumi: tt.args.Bangumi,
			}
			if tt.playlist {
				// playlist mode
				config.Playlist = true
				Download(tt.args.URL)
				// single mode
				config.Playlist = false
				Download(tt.args.URL)
				data = bilibiliDownload(tt.args.URL, options)
			} else {
				config.Playlist = false
				data = bilibiliDownload(tt.args.URL, options)
			}
			test.Check(t, tt.args, data)
		})
	}
}
