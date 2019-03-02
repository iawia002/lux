package douyin

import (
	"testing"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/test"
)

func TestDownload(t *testing.T) {
	config.InfoOnly = true
	config.RetryTimes = 100
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "normal test",
			args: test.Args{
				URL:   "https://www.douyin.com/share/video/6557825773007277319/?mid=6557826301539912456",
				Title: "跟特效师一起学跳舞，看变形金刚擎天柱怎么跳，你也来试试！@抖音小助手",
				Size:  4927877,
			},
		},
		{
			name: "no title test",
			args: test.Args{
				URL:   "http://v.douyin.com/Yy76ML/",
				Title: "抖音短视频",
				Size:  4320955,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := Extract(tt.args.URL)
			test.CheckError(t, err)
			test.Check(t, tt.args, data[0])
		})
	}
}
