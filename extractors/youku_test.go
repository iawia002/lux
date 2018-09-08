package extractors

import (
	"testing"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/test"
)

func TestYouku(t *testing.T) {
	config.InfoOnly = true
	config.Ccode = "0103010102"
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "normal test",
			args: test.Args{
				URL:     "http://v.youku.com/v_show/id_XMzUzMjE3NDczNg==.html",
				Title:   "车事儿：智能汽车已经不在遥远 东风风光iX5发布",
				Size:    45185427,
				Quality: "mp4hd3 1920x1080",
			},
		},
		{
			name: "normal test",
			args: test.Args{
				URL:     "http://v.youku.com/v_show/id_XMzQ1MTAzNjQwNA==.html",
				Title:   "这！就是街舞 第一季 第3期：百强“互杀”队长不忍直视",
				Size:    1419459808,
				Quality: "mp4hd3 1920x1080 国语",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, _ := Youku(tt.args.URL)
			test.Check(t, tt.args, data)
		})
	}
}
