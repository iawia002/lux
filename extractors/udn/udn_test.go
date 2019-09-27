package udn

import (
	"testing"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/test"
)

func TestExtract(t *testing.T) {
	config.InfoOnly = true
	config.RetryTimes = 10
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "normal test",
			args: test.Args{
				URL:   "https://video.udn.com/embed/news/300040",
				Title: `生物老師男變女 全校挺"做自己"`,
				Size:  12740874,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := Extract(tt.args.URL)
			test.CheckError(t, err)
			test.Check(t, tt.args, data[0])
		})
	}
}
