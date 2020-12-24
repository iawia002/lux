package ixigua

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
				URL:   "https://www.ixigua.com/6909412821625209357",
				Title: "赌场上十赌九输，想知道你赢不了的原因吗？这部电影告诉你答案！",
			},
		}, {
			name: "episode test",
			args: test.Args{
				URL:   "https://www.ixigua.com/6841411932574974477?id=6844443142310068749",
				Title: "觉醒 第16集",
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
