package netease

import (
	"testing"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/test"
)

func TestDownload(t *testing.T) {
	config.InfoOnly = true
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "mv test 1",
			args: test.Args{
				URL:   "https://music.163.com/#/mv?id=5547010",
				Title: "There For You - Troye Sivan - 高清MV - 网易云音乐",
				Size:  24249078,
			},
		},
		{
			name: "video test 1",
			args: test.Args{
				URL:   "https://music.163.com/#/video?id=C8C9D11629798595BD28451DE3AC9FF4",
				Title: "＃金曜日の新垣结衣 总集編〈全9編〉 - 视频 - 网易云音乐",
				Size:  37408123,
			},
		},
		{
			name: "video test 2",
			args: test.Args{
				URL:   "https://music.163.com/m/video?id=6309CF62EF5D44FED5974536604944CF&userid=567080617",
				Title: "当皮卡丘失去了小智就失去了全世界 - 视频 - 网易云音乐",
				Size:  28547736,
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
