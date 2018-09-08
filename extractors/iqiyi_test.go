package extractors

import (
	"testing"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/test"
)

func TestIqiyi(t *testing.T) {
	config.InfoOnly = true
	config.RetryTimes = 100
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
				URL:     "http://www.iqiyi.com/v_19rqy2z83w.html",
				Title:   "收了创意视频2018：58天环球飞行记",
				Size:    41235168,
				Quality: "1280x720",
			},
		},
		{
			name: "curid test 1",
			args: test.Args{
				URL:     "https://www.iqiyi.com/v_19rro0jdls.html#curid=350289100_6e6601aae889d0b1004586a52027c321",
				Title:   "Shawn Mendes - Never Be Alone",
				Size:    41084204,
				Quality: "1280x528",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, _ := Iqiyi(tt.args.URL)
			test.Check(t, tt.args, data)
		})
	}
}
