package yinyuetai

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
			name: "normal test",
			args: test.Args{
				URL:     "http://v.yinyuetai.com/video/3310345",
				Title:   "周杰伦 - 七里香",
				Size:    118380541,
				Quality: "超清",
			},
		},
		{
			name: "h5 test",
			args: test.Args{
				URL:     "http://v.yinyuetai.com/video/h5/820981",
				Title:   "Rap God",
				Size:    144401919,
				Quality: "超清",
			},
		},
		{
			name: "mobile test",
			args: test.Args{
				URL:     "http://m2.yinyuetai.com/video.html?id=3310363",
				Title:   "周杰伦 - 等你下课",
				Size:    105821486,
				Quality: "超清",
			},
		},
		{
			name: "normal test special",
			args: test.Args{
				URL:     "http://v.yinyuetai.com/video/3164150?vid=3164151",
				Title:   "Lemon",
				Size:    15993202,
				Quality: "流畅",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := Download(tt.args.URL)
			test.CheckError(t, err)
			test.Check(t, tt.args, data[0])
		})
	}
}
