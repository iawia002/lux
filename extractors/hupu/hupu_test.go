package hupu

import (
	"testing"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/test"
)

func TestHupu(t *testing.T) {
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "normal test",
			args: test.Args{
				URL:   "https://bbs.hupu.com/47401018.html?is_reflow=1&cid=84752419&bddid=56KXU5QUJH4VGM26SFPTYTKNI5CFNJMX736TIZ52DXLGUAAMBJVA01&puid=16522089&client=8577E496-4D9B-4E5C-A9DB-A8EF5C1956D2",
				Title: "结局引起舒适",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			New().Extract(tt.args.URL, extractors.Options{})
		})
	}
}
