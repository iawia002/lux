package tiktok

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
			name: "normal test 1",
			args: test.Args{
				URL:   "https://www.tiktok.com/@ginjiro_koyama/video/7164293510617763073?is_copy_url=1&is_from_webapp=v1",
				Title: "イケすぎたXOXO#xoxo #repezenfoxx #背中男 #kfam #yoshikiさんを泣かせたチーム @K fam @【Repezen Foxx】🦊",
				Size:  4356253,
			},
		},
		{
			name: "normal test 2",
			args: test.Args{
				URL:   "https://www.tiktok.com/@enhypen/video/7165445991238356225?is_copy_url=1&is_from_webapp=v1",
				Title: "깜짝 퇴장 👋 #ENHYPEN #SUNGHOON #NI_KI #Make_the_change",
				Size:  3848307,
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
