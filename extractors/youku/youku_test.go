package youku

import (
	"testing"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/test"
)

func TestDownload(t *testing.T) {
	config.InfoOnly = true
	config.RetryTimes = 100
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
		{
			name: "normal test",
			args: test.Args{
				URL:     "http://v.youku.com/v_show/id_XMzQ1MTAzNjQwNA==.html",
				Title:   "这！就是街舞 第一季 百强“互杀”队长不忍直视，黄子韬组内上演街舞“世纪大战”",
				Size:    750911635,
				Quality: "mp4hd2v2 1280x720 国语",
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
