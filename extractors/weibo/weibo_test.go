package weibo

import (
	"testing"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/test"
)

func TestDownload(t *testing.T) {
	config.InfoOnly = true
	config.RetryTimes = 10
	tests := []struct {
		name string
		args test.Args
	}{
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
				Title: `近日，日本视错觉大师、明治大学特任教授\"杉原厚吉的“错觉箭头“作品又引起世界人民的关注。反射，透视和视角的巧妙结合产生了这种惊人的幻觉：箭头向右？转过来还是向右？\n\n引用杉原教授的经典描述：“我们看外面的世界的方式——也就是我们的知觉——都是由大脑机制间接产生的，所以所有知觉在某`,
				Size:  1125984,
			},
		},
		{
			name: "weibo.com test",
			args: test.Args{
				URL:   "https://weibo.com/1642500775/GjbO5ByzE",
				Title: "让人怦然心动的小姐姐们 via@大懒糖",
				Size:  2002420,
			},
		},
		{
			name: "weibo.com/tv test",
			args: test.Args{
				URL:     "https://weibo.com/tv/v/jGz6llNZ1?fid=1034:4298353237002268",
				Title:   "做了这么一个屌炸天的视频我也不知道起什么标题好 @DRock-Art @毒液-致命守护者 @漫威影业 #绘画# #blender# #漫威#",
				Quality: "720",
				Size:    7520929,
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
