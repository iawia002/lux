package youku

import (
	"testing"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/test"
)

func TestDownload(t *testing.T) {
	config.InfoOnly = true
	config.RetryTimes = 10
	config.YoukuCcode = "0590"
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "normal test",
			args: test.Args{
				URL:     "http://v.youku.com/v_show/id_XMzUzMjE3NDczNg==.html",
				Title:   "车事儿: 智能汽车已经不在遥远 东风风光iX5发布",
				Size:    22692900,
				Quality: "mp4hd2v2 1280x720",
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
