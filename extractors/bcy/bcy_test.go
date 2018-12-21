package bcy

import (
	"testing"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/test"
)

func TestDownload(t *testing.T) {
	config.InfoOnly = true
	config.RetryTimes = 100
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "normal test",
			args: test.Args{
				URL:   "https://bcy.net/illust/detail/38134/2048276",
				Title: "牛奶小姐姐，草莓味的w",
				Size:  2647838,
			},
		},
		{
			name: "normal test",
			args: test.Args{
				URL:   "https://bcy.net/coser/detail/143767/2094010",
				Title: "phx：柠檬先行预告！牧濑红莉栖 cn: 三度",
				Size:  3329959,
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
