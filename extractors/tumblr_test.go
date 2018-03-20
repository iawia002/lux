package extractors

import (
	"testing"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/test"
)

func TestTumblr(t *testing.T) {
	config.InfoOnly = true
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "normal test",
			args: test.Args{
				URL:   "http://fuckyeah-fx.tumblr.com/post/170392654141/180202-%E5%AE%8B%E8%8C%9C",
				Title: "f(x)",
				Size:  1690025,
			},
		},
		{
			name: "normal test",
			args: test.Args{
				URL:   "http://therealautoblog.tumblr.com/post/171623222197/paganis-new-projects-huayra-successor-with",
				Title: "Autoblog • Pagani’s new projects: Huayra successor with...",
				Size:  154722,
			},
		},
		{
			name: "normal test",
			args: test.Args{
				URL:   "https://outdoorspastelnature.tumblr.com/post/170380315768/feel-at-peace",
				Title: "Pastel Nature — Feel at Peace",
				Size:  514444,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := Tumblr(tt.args.URL)
			test.Check(t, tt.args, data)
		})
	}
}
