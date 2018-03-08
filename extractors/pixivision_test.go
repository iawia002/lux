package extractors

import (
	"testing"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/test"
)

func TestPixivision(t *testing.T) {
	config.InfoOnly = true
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "normal test",
			args: test.Args{
				URL:   "https://www.pixivision.net/zh/a/3271",
				Title: "Don't ask me to choose! Tiny Breasts VS Huge Breasts",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := Pixivision(tt.args.URL)
			test.Check(t, tt.args, data)
		})
	}
}
