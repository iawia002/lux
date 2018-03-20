package extractors

import (
	"testing"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/test"
)

func TestVimeo(t *testing.T) {
	config.InfoOnly = true
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "normal test",
			args: test.Args{
				URL:     "https://player.vimeo.com/video/259325107",
				Title:   "prfm 20180309",
				Size:    131051118,
				Quality: "1080p",
			},
		},
		{
			name: "normal test",
			args: test.Args{
				URL:     "https://vimeo.com/254865724",
				Title:   "MAGIC DINER PT. II",
				Size:    138966306,
				Quality: "1080p",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := Vimeo(tt.args.URL)
			test.Check(t, tt.args, data)
		})
	}
}
