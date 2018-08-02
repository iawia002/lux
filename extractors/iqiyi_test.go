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
				URL:     "http://www.iqiyi.com/v_19rrbdmaj0.html",
				Title:   "新一轮降水将至 冷空气影响中东部地区",
				Size:    3122304,
				Quality: "896x504",
			},
		},
		{
			name: "title test 1",
			args: test.Args{
				URL:     "http://www.iqiyi.com/v_19rrbhikxo.html",
				Title:   "热血街舞团：鹿晗个人宣传片震撼发布 执着前行终现万丈荣光",
				Size:    13843192,
				Quality: "1280x720",
			},
		},
		{
			name: "curid test 1",
			args: test.Args{
				URL:     "http://www.iqiyi.com/v_19rrbebyt8.html#curid=963226000_f877de17f261458b04c932ca90ee67a3",
				Title:   "热血街舞团：召集人公演炸屏来袭 热血之城上演“抢人”大战",
				Size:    344675816,
				Quality: "3840x2160",
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
