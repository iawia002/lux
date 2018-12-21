package instagram

import (
	"testing"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/test"
)

func TestDownload(t *testing.T) {
	config.InfoOnly = true
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "video test",
			args: test.Args{
				URL:   "https://www.instagram.com/p/BlIka1ZFCNr",
				Title: "P!NK on Instagram: “AFL got us hyped! #adelaideadventures #iwanttoplay”",
				Size:  2741413,
			},
		},
		{
			name: "image test",
			args: test.Args{
				URL:   "https://www.instagram.com/p/Bl5oVUyl9Yx",
				Title: "P!NK on Instagram: “Australia:heaven”",
				Size:  250596,
			},
		},
		{
			name: "image album test",
			args: test.Args{
				URL:   "https://www.instagram.com/p/Bjyr-gxF4Rb",
				Title: "P!NK on Instagram: “Nature. Nurture.\nKiddos. Gratitude”",
				Size:  4599909,
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
