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
			name: "vid test",
			args: test.Args{
				URL:     "https://v.qq.com/x/cover/4opd3z8rcb7bbhh.html",
				Title:   "《古墓丽影：源起之战》中文版终极预告 “坎妹”吴彦祖爆燃开打",
				Quality: "蓝光;(1080P)",
			},
		},
		{
			name: "movie test",
			args: test.Args{
				URL:     "https://v.qq.com/x/cover/e5qmd3z5jr0uigk/9k9DQFMFRoo.html",
				Title:   "赌侠",
				Size:    470753231,
				Quality: "高清;(480P)",
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
