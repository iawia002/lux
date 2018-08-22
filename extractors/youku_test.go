package extractors

import (
	"testing"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/test"
)

func init() {
	config.Ccode = "010101500003"
}

func TestYouku(t *testing.T) {
	config.InfoOnly = true
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "normal test",
			args: test.Args{
				URL:     "http://v.youku.com/v_show/id_XMzUzMjE3NDczNg==.html",
				Title:   "车事儿：智能汽车已经不在遥远 东风风光iX5发布",
				Size:    22692900,
				Quality: "mp4hd2 1280x720",
			},
		},
		{
			name: "normal test",
			args: test.Args{
				URL:     "http://v.youku.com/v_show/id_XMzQ1MTAzNjQwNA==.html",
				Title:   "这！就是街舞 第一季 第3期：百强“互杀”队长不忍直视",
				Size:    750911635,
				Quality: "mp4hd2 1280x720 国语",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := Youku(tt.args.URL)
			test.Check(t, tt.args, data)
		})
	}
}
