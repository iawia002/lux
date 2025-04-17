package rumble

import (
	"testing"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/test"
)

func TestRumble(t *testing.T) {
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "normal test",
			args: test.Args{
				URL:   "https://rumble.com/v24swn0-just-say-yes-to-climate-lockdowns.html",
				Title: "Just Say YES to Climate Lockdowns!",
			},
		},
		{
			name: "normal test",
			args: test.Args{
				URL:   "https://rumble.com/v6rmfm1-monday-full-show-33125-hhs-head-rfk-jr.-pledges-to-stopv.html",
				Title: "MONDAY FULL SHOW 3/31/25 — HHS Head RFK Jr. Pledges To Stopv",
			},
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := New().Extract(tt.args.URL, extractors.Options{})
			if err != nil {
				t.Error(err)
			}
			for _, d := range data {
				found := false
				for _, s := range d.Streams {
					if s.Size > 0 {
						found = true
					}
				}
				if !found {
					t.Errorf("no streams found in test %d", i)
				}
			}
		})
	}
}
