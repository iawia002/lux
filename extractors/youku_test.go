package extractors

import (
	"testing"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/test"
)

func TestYouku(t *testing.T) {
	config.InfoOnly = true
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "normal test",
			args: test.Args{
				URL:     "http://v.youku.com/v_show/id_XMzMzMDk5MzcyNA==.html",
				Title:   "鲜榨时尚：开口跪Jessie J歌声和造型双在线",
				Size:    12568063,
				Quality: "mp4hd2v2 720x1280",
			},
		},
		{
			name: "normal test",
			args: test.Args{
				URL:     "http://v.youku.com/v_show/id_XMzQ1MTAzNjQwNA==.html",
				Title:   "这！就是街舞 第一季：第3期：百强“互杀”队长不忍直视",
				Size:    806126651,
				Quality: "mp4hd2 1280x720",
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
