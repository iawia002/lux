package tiktok

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
			name: "normal test 1",
			args: test.Args{
				URL:   "https://www.tiktok.com/@therock/video/6768158408110624005",
				Title: "#bestfriend check.",
				Size:  2594827,
			},
		},
		{
			name: "normal test 2",
			args: test.Args{
				URL:   "https://www.tiktok.com/@yun_bao/video/7050411198512155905",
				Title: "🤔",
				Size:  2224436,
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
