package streamtape

import (
	"testing"

	"github.com/iawia002/lux/extractors/types"
	"github.com/iawia002/lux/test"
)

func TestStreamtape(t *testing.T) {
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "normal test",
			args: test.Args{
				URL:   "https://streamtape.com/e/vkoKlwYPo9F4mRo",
				Title: "lux.mp4",
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
