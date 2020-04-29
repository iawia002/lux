package qingting

import (
	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/test"
	"testing"
)

func TestExtract(t *testing.T) {
	config.InfoOnly = true
	config.ThreadNumber = 9
	tests := []struct {
		name     string
		args     test.Args
		playlist bool
	}{
		{
			name: "playlist test",
			args: test.Args{
				URL:   "https://www.qingting.fm/channels/226572",
				Title: "ViliBili | 这个冬天是个恋爱的季节",
				Size:  66284484,
			},
			playlist: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				data []downloader.Data
				err  error
			)
			if tt.playlist {
				// playlist mode
				config.Playlist = true
				_, err = Extract(tt.args.URL)
				test.CheckError(t, err)
			} else {
				config.Playlist = false
				data, err = Extract(tt.args.URL)
				test.CheckError(t, err)
				test.Check(t, tt.args, data[0])
			}
		})
	}
}
