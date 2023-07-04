package pinterest

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
				URL:   "https://www.pinterest.com/pin/creamy-cheesy-pretzel-bites-video--368450813272292084/",
				Title: "Creamy Cheesy Pretzel Bites [Video] ",
				Size:  30247497,
			},
		},
		{
			name: "normal test 2",
			args: test.Args{
				URL:   "https://www.pinterest.com/pin/532198880988430823/",
				Title: "Pin on TikTok ~ The world of food",
				Size:  4676927,
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
