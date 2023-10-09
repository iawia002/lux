package zingmp3

import (
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
			name: "Host is mp3.zing.vn",
			args: test.Args{
				URL:   "https://mp3.zing.vn/bai-hat/Xa-Mai-Xa-Bao-Thy/ZWZB9WAB.html",
				Title: "Xa Mãi Xa",
			},
		},
		{
			name: "Host is zingmp3.vn",
			args: test.Args{
				URL:   "https://zingmp3.vn/bai-hat/SOLO-JENNIE/ZW9FID6Z.html",
				Title: "SOLO",
			},
		},
		{
			name: "Video clip",
			args: test.Args{
				URL:   "https://zingmp3.vn/video-clip/Suong-Hoa-Dua-Loi-K-ICM-RYO/ZO8ZF7C7.html",
				Title: "Sương Hoa Đưa Lối",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := New().Extract(tt.args.URL, extractors.Options{})
			test.CheckError(t, err)
			test.Check(t, tt.args, data[0])
		})
	}
}
