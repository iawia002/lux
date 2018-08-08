package extractors

import (
	"testing"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/test"
)

func TestQQ(t *testing.T) {
	config.InfoOnly = true
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "normal test",
			args: test.Args{
				URL:     "https://v.qq.com/x/page/n0687peq62x.html",
				Title:   "世界杯第一期：100秒速成！“伪球迷”世界杯生存指南",
				Size:    23897406,
				Quality: "蓝光;(1080P)",
			},
		},
		{
			name: "movie and vid test",
			args: test.Args{
				URL:     "https://v.qq.com/x/cover/e5qmd3z5jr0uigk.html",
				Title:   "赌侠1991（普通话版）",
				Size:    470753231,
				Quality: "高清;(480P)",
			},
		},
		{
			name: "single part test",
			args: test.Args{
				URL:     "https://v.qq.com/iframe/player.html?vid=v0739eolv38",
				Title:   "PGI国际邀请赛，FPP第四局，OMG强势吃鸡，全场观众高喊OMG",
				Size:    10714773,
				Quality: "标清;(270P)",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := QQ(tt.args.URL)
			test.Check(t, tt.args, data)
		})
	}
}
