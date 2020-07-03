package weibo

import (
	"testing"

	"github.com/iawia002/annie/extractors/types"
	"github.com/iawia002/annie/test"
)

func TestToken(t *testing.T) {
	t.Run(
		"XSRF token test", func(t *testing.T) { getXSRFToken() },
	)
}

func TestDownload(t *testing.T) {
	tests := []struct {
		name string
		args test.Args
	}{
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
				URL:     "https://weibo.com/tv/show/1034:4298353237002268?from=old_pc_videoshow",
				Title:   "毒液插图Blender+Photoshop2.5小时工作流",
				Quality: "720p",
				Size:    7520929,
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
