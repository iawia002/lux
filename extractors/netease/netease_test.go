package netease

import (
	"testing"

	"github.com/iawia002/annie/extractors/types"
	"github.com/iawia002/annie/test"
)

func TestDownload(t *testing.T) {
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "mv test 1",
			args: test.Args{
				URL:   "https://music.163.com/#/mv?id=5547010",
				Title: "There For You",
				Size:  24249078,
			},
		},
		{
			name: "video test 1",
			args: test.Args{
				URL:   "https://music.163.com/#/video?id=C8C9D11629798595BD28451DE3AC9FF4",
				Title: "＃金曜日の新垣结衣 总集編〈全9編〉",
				Size:  37408123,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := New().Extract(tt.args.URL, types.Options{})
			test.CheckError(t, err)
			test.Check(t, tt.args, data[0])
		})
	}
}
