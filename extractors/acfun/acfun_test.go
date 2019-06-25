package acfun

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
			name: "normal test 1",
			args: test.Args{
				URL:     "https://www.acfun.cn/v/ac10277559",
				Title:   "全套ps抠图视频教程！1-1：使用【矩形选框工具】抠方形图像",
				Quality: "960x664",
				Size:    5673729,
			},
		},
		{
			name: "normal test 2",
			args: test.Args{
				URL:     "https://www.acfun.cn/v/ac10308664",
				Title:   "【MMD教程】如何快速做出一个跑步【K帧教程】Part1",
				Quality: "1920x1080",
				Size:    705043229,
			},
		},
		{
			name: "normal test 3",
			args: test.Args{
				URL:     "https://www.acfun.cn/v/ac4074170",
				Title:   "AB向：香港电影中的经典舞蹈（二）Part1",
				Quality: "1920x1080",
				Size:    222034640,
			},
		},
		{
			name: "bangumi test 1",
			args: test.Args{
				URL:     "https://www.acfun.cn/bangumi/ab5020318_29434_260696",
				Title:   "Lovelive! Sunshine!! 第二季第13话",
				Quality: "1920x1080",
				Size:    591260242,
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
