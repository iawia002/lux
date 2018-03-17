package extractors

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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data downloader.VideoData
			if tt.playlist {
				// playlist mode
				config.Playlist = true
				Bilibili(tt.args.URL)
				// single mode
				config.Playlist = false
				data = bilibiliDownload(tt.args.URL, tt.args.Bangumi)
			} else {
				config.Playlist = false
				data = bilibiliDownload(tt.args.URL, tt.args.Bangumi)
			}
			test.Check(t, tt.args, data)
		})
	}
}
