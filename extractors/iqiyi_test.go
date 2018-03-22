package extractors

import (
	"testing"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/test"
)

func TestIqiyi(t *testing.T) {
	config.InfoOnly = true
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "normal test",
			args: test.Args{
				URL:     "http://www.iqiyi.com/v_19rrbhikxo.html",
				Title:   "热血街舞团：鹿晗个人宣传片震撼发布 执着前行终现万丈荣光",
				Size:    12750912,
				Quality: "1280x720",
			},
		},
		{
			name: "normal test",
			args: test.Args{
				URL:     "http://www.iqiyi.com/v_19rrbdmaj0.html",
				Title:   "新一轮降水将至 冷空气影响中东部地区",
				Size:    2838236,
				Quality: "896x504",
			},
		},
		{
			name: "normal test",
			args: test.Args{
				URL:     "http://www.iqiyi.com/a_19rrhcqtot.html#curid=958070800_e05591c8ad96022f79f41ec4fcc611a9",
				Title:   "《热血街舞团》综艺专题-高清视频在线观看-爱奇艺",
				Size:    12750912,
				Quality: "1280x720",
			},
		},
		{
			name: "normal test",
			args: test.Args{
				URL:     "http://www.iqiyi.com/a_19rrhcqtot.html#curid=958065800_03b77bd0648a6c1df86b0f7c4fd0e526",
				Title:   "《热血街舞团》综艺专题-高清视频在线观看-爱奇艺",
				Size:    7743532,
				Quality: "1280x720",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := Iqiyi(tt.args.URL)
			test.Check(t, tt.args, data)
		})
	}
}
