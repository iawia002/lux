package extractors

import (
	"testing"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/test"
)

func TestTwitter(t *testing.T) {
	config.InfoOnly = true
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "normal test",
			args: test.Args{
				URL:     "https://twitter.com/justinbieber/status/898217160060698624",
				Title:   "Justin Bieber 898217160060698624",
				Size:    1733172,
				Quality: "720x1280",
			},
		},
		{
			name: "abnormal uri test1",
			args: test.Args{
				URL:     "https://twitter.com/twitter/statuses/898567934192177153",
				Title:   "Justin Bieber 898567934192177153",
				Size:    3985600,
				Quality: "1280x720",
			},
		},
		{
			name: "abnormal uri test2",
			args: test.Args{
				URL:     "https://twitter.com/kyoudera/status/971819131711373312/video/1/",
				Title:   "ネメシス 京寺 971819131711373312",
				Size:    13941892,
				Quality: "1280x720",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := Twitter(tt.args.URL)
			test.Check(t, tt.args, data)
		})
	}
}
