package bitchute

import (
	"testing"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/test"
)

func TestDownload(t *testing.T) {
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "video test 1",
			args: test.Args{
				URL:   "https://www.bitchute.com/video/C17naZ6WlWPo",
				Title: "Everybody Dance Now",
				Size:  1794720,
			},
		},
		{
			name: "video test 2",
			args: test.Args{
				URL:   "https://www.bitchute.com/video/HFgoUz6HrvQd",
				Title: "Bear Level 1",
				Size:  971698,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := New().Extract(tt.args.URL, extractors.Options{})
			test.CheckError(t, err)
			test.Check(t, tt.args, data[0])
		})
	}
}
