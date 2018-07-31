package extractors

import (
	"testing"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/test"
)

func TestLe(t *testing.T) {
	config.InfoOnly = true
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "normal test",
			args: test.Args{
				URL:     "http://www.le.com/ptv/vplay/31453073.html",
				Title:   "难以管理！中国共享单车撤出华盛顿",
				Size:    7006196,
				Quality: "高清 960x544",
			},
		},
		{
			name: "comic test",
			args: test.Args{
				URL:     "http://www.le.com/ptv/vplay/31448498.html",
				Title:   "天行九歌 60",
				Size:    398149972,
				Quality: "1080P 1920x1072",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := Le(tt.args.URL)
			test.Check(t, tt.args, data)
		})
	}
}
