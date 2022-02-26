package tangdou

import (
	"testing"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/test"
)

func TestTangDou(t *testing.T) {
	tests := []struct {
		name     string
		args     test.Args
		playlist bool
	}{
		{
			name: "need call share url first and get the signed video URL test and can get title from head's title tag",
			args: test.Args{
				URL:   "https://m.tangdou.com/play/1500676338077",
				Title: "暴瘦减肚子，不用跑不用跳，8天瘦了16斤 正面演示 背面演示 分解教学__广场舞_糖豆广场舞-糖豆视频",
				Size:  62258444,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				data []*extractors.Data
				err  error
			)
			if tt.playlist {
				// playlist mode
				_, err = New().Extract(tt.args.URL, extractors.Options{
					Playlist:     true,
					ThreadNumber: 9,
				})
				test.CheckError(t, err)
			} else {
				data, err = New().Extract(tt.args.URL, extractors.Options{})
				test.CheckError(t, err)
				test.Check(t, tt.args, data[0])
			}
		})
	}
}
