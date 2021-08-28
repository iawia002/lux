package douyin

import (
	"testing"

	"github.com/iawia002/annie/extractors/types"
	"github.com/iawia002/annie/test"
)

func TestDownload(t *testing.T) {
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "normal test",
			args: test.Args{
				URL:   "https://www.douyin.com/video/6967223681286278436?previous_page=main_page&tab_name=home",
				Title: "是爱情，让父子相认#陈翔六点半  #关于爱情",
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
