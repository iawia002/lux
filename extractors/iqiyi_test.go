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
				Size:    11718040,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := Iqiyi(tt.args.URL)
			test.Check(t, tt.args, data)
		})
	}
}
