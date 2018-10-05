package weibo

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
				URL:   "https://m.weibo.cn/2815133121/G9VBqbsWM",
				Title: "当你超过25岁再去夜店……",
				Size:  3112080,
			},
		},
		{
			name: "fid url test",
			args: test.Args{
				URL:   "https://weibo.com/tv/v/Ga7XazXze?fid=1034:4a65c6e343dc672789d3ba49c2463c6a",
				Title: "看完更加睡不着了[二哈]",
				Size:  438757,
			},
		},
		{
			name: "title test",
			args: test.Args{
				URL:   "https://m.weibo.cn/status/4237529215145705",
				Title: `近日，日本视错觉大师、明治大学特任教授\"杉原厚吉的“错觉箭头“作品又引起世界人民的关注。反射，透视和视角的巧妙结合产生了这种惊人的幻觉：箭头向右？转过来...`,
				Size:  1125984,
			},
		},
		{
			name: "weibo.com test",
			args: test.Args{
				URL:   "https://weibo.com/1642500775/GjbO5ByzE",
				Title: "让人怦然心动的小姐姐们 via@大懒糖",
				Size:  9198410,
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
