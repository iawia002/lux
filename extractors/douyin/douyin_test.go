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
				URL:   "https://v.douyin.com/RHanxuR/",
				Title: "树叶只有树 树却有很多树叶",
			},
		},
		{
			name: "image test",
			args: test.Args{
				URL:   "https://v.douyin.com/RHanxuR/",
				Title: "树叶只有树 树却有很多树叶",
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
