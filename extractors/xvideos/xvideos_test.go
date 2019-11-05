package xvideos

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
				URL:   "https://www.xvideos.com/video29018757/asian_chick_enjoying_sex_debut._hd_full_at_nanairo.co",
				Title: "Asian chick enjoying sex debut&period; HD FULL at&colon; nanairo&period;co - XVIDEOS.COM",
				Size:  16574766,
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
