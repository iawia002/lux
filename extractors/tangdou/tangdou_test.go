package tangdou

import (
	"testing"

	"github.com/iawia002/annie/extractors/types"
	"github.com/iawia002/annie/test"
)

func TestTangDou(t *testing.T) {
	tests := []struct {
		name     string
		args     test.Args
		playlist bool
	}{
		{
			name: "contains video URL test directly and can get title from body's div tag",
			args: test.Args{
				URL:   "http://www.tangdou.com/v95/dAOQNgMjwT2D5w2.html",
				Title: "杨丽萍广场舞《好日子天天过》喜庆双扇扇子舞",
			},
		},
		{
			name: "need call share url first and get the signed video URL test and can get title from head's title tag",
			args: test.Args{
				URL:   "http://m.tangdou.com/v94/dAOMMYNjwT1T2Q2.html",
				Title: "吉美广场舞《再唱山歌给党听》民族形体舞 附教学视频在线观看",
				Size:  50710318,
			},
		},
		{
			name: "playlist test",
			args: test.Args{
				URL:   "http://www.tangdou.com/playlist/view/2816/page/4",
				Title: "茉莉广场舞 我向草原问个好 原创藏族风民族舞附教学",
				Size:  66284484,
			},
			playlist: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				data []*types.Data
				err  error
			)
			if tt.playlist {
				// playlist mode
				_, err = New().Extract(tt.args.URL, types.Options{
					Playlist:     true,
					ThreadNumber: 9,
				})
				test.CheckError(t, err)
			} else {
				data, err = New().Extract(tt.args.URL, types.Options{})
				test.CheckError(t, err)
				test.Check(t, tt.args, data[0])
			}
		})
	}
}
