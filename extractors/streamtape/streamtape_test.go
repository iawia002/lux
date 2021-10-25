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
			data, err := New().Extract(tt.args.URL, types.Options{})
			if err != nil {
				t.Error(err)
				return
			}
			if len(data) == 0 {
				t.Error("extractor returned empty data")
				return
			}
			if data[0].Title != tt.args.Title {
				t.Errorf("expected title '%s' got '%s'", tt.args.Title, data[0].Title)
			}
		})
	}
}
