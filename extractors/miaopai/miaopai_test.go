package miaopai

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
				URL:   "https://www.miaopai.com/show/nPWJvdR4z2Bg1Sz3PJpNYffjpDgEiuv4msALgw__.htm",
				Title: "情人节特辑：一个来自绝地求生的爱情故事，送给已经离开的你",
				Size:  15756710,
			},
		},
		{
			name: "normal test 2",
			args: test.Args{
				URL:   "https://m.miaopai.com/show/channel/3PCuI5IZ6wdSZISmTtasYTa-l~wrVxk1yEgWRQ__",
				Title: "小学霸6点半起床学习:想赢在起跑线",
				Size:  8783794,
			},
		},
		{
			name: "normal test 3",
			args: test.Args{
				URL:   "http://n.miaopai.com/media/qVWj3dVK2oSxtW~vSq2tGeBKPE--tPSp",
				Title: "如果你家的猫喜欢咬人怎么办？",
				Size:  4794459,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := Download(tt.args.URL)
			test.CheckError(t, err)
			test.Check(t, tt.args, data[0])
		})
	}
}
