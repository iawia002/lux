package huya

import (
	"testing"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/test"
)

func TestHuya(t *testing.T) {
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "normal test",
			args: test.Args{
				URL:   "https://m.v.huya.com/play/fans/630103747.html/?shareid=4597484513543964249&shareUid=2179142017&source=ios&sharetype=other&platform=2",
				Title: "12.28 集梦薛小谦【封号斗罗】直播名场面",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			New().Extract(tt.args.URL, extractors.Options{})
		})
	}
}
