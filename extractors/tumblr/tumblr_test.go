package tumblr

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
			name: "image test 1",
			args: test.Args{
				URL:   "http://fuckyeah-fx.tumblr.com/post/170392654141/180202-%E5%AE%8B%E8%8C%9C",
				Title: "f(x)",
			},
		},
		{
			name: "image test 2",
			args: test.Args{
				URL:   "http://therealautoblog.tumblr.com/post/171623222197/paganis-new-projects-huayra-successor-with",
				Title: "Autoblog • Pagani’s new projects: Huayra successor with...",
			},
		},
		{
			name: "video test",
			args: test.Args{
				URL:   "https://boomgoestheprower.tumblr.com/post/174127507696",
				Title: "See, that’s what the app is perfect for.",
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
