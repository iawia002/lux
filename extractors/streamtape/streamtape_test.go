package streamtape

import (
	"testing"

	"github.com/iawia002/annie/extractors/types"
	"github.com/iawia002/annie/test"
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
				Title: "annie.mp4",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := New().Extract(tt.args.URL, types.Options{}); err != nil {
				t.Error(err)
			}
		})
	}
}
