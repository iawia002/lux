package kuaishou

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
			name: "normal test",
			args: test.Args{
				URL:     "https://www.kuaishou.com/short-video/3x43cyvcyph57i4?authorId=3xtq3uqyjmhbimq&streamSource=find&area=homexxbrilliant",
				Title:   "现在连戴口罩都开始内卷了吗？！快get口罩心机戴法，直接戴出小V脸啊 ！ #口罩 #显脸小-快手",
				Size:    1077774,
				Quality: "sd",
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
