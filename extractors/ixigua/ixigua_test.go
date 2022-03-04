package ixigua

import (
	"log"
	"testing"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/test"
)

func TestDownload(t *testing.T) {
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "test 1",
			args: test.Args{
				URL:   "https://www.ixigua.com/7053389963487871502",
				Title: "漫威斥巨资拍的《永恒族》，刚上架就被多国禁播，究竟拍了什么？",
			},
		},
		{
			name: "test 2",
			args: test.Args{
				URL:   "https://v.ixigua.com/RedcbWM/",
				Title: "为长生不老，竟然连小鲛人都杀 @中视频伙伴计划官号",
			},
		},
		{
			name: "test 3",
			args: test.Args{
				URL:   "https://m.toutiao.com/is/dtj1pND/",
				Title: "卡尔：59杀4200法强小法师，点塔只需一下，W技能瞬秒对方",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log.Println(tt.args.URL)
			New().Extract(tt.args.URL, extractors.Options{})
		})
	}
}
