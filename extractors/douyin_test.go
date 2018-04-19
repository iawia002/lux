package extractors

import (
	"testing"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/test"
)

func TestDouyin(t *testing.T) {
	config.InfoOnly = true
	tests := []struct {
		name string
		args test.Args
	}{
		// {
		// 	name: "normal test",
		// 	args: test.Args{
		// 		URL:   "https://www.douyin.com/share/video/6509219899754155272",
		// 		Title: "好冷  逢考必过",
		// 	},
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := Douyin(tt.args.URL)
			test.Check(t, tt.args, data)
		})
	}
}
