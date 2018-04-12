package extractors

import (
	"testing"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/test"
)

func TestWeibo(t *testing.T) {
	config.InfoOnly = true
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "normal test",
			args: test.Args{
				URL:   "https://m.weibo.cn/2815133121/G9VBqbsWM",
				Title: "当你超过25岁再去夜店……",
				Size:  3112080,
			},
		},
		{
			name: "normal test",
			args: test.Args{
				URL:   "https://weibo.com/tv/v/Ga7XazXze?fid=1034:4a65c6e343dc672789d3ba49c2463c6a",
				Title: "看完更加睡不着了[二哈]",
				Size:  438757,
			},
		},
		{
			name: "normal test",
			args: test.Args{
				URL:   "https://m.weibo.cn/status/4226449584043915",
				Title: "荷兰弟根本不知道跟谁演的对手戏 via@谷大白话",
				Size:  30429685,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := Weibo(tt.args.URL)
			test.Check(t, tt.args, data)
		})
	}
}
