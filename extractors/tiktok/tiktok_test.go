package tiktok

import (
	"testing"

	"github.com/iawia002/annie/extractors/types"
	"github.com/iawia002/annie/test"
)

func TestDownload(t *testing.T) {
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "normal test",
			args: test.Args{
				URL:   "https://www.tiktok.com/@therock/video/6768158408110624005",
				Title: "#bestfriend check.",
			},
		},
		{
			name: "short url test",
			args: test.Args{
				URL:   "https://vm.tiktok.com/C998PY/",
				Title: "Who saw that coming? 🍁 #leaves #fall",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := New().Extract(tt.args.URL, types.Options{})
			test.CheckError(t, err)
			test.Check(t, tt.args, data[0])
		})
	}
}
