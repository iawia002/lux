package instagram

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
			name: "video test",
			args: test.Args{
				URL:   "https://www.instagram.com/p/BlIka1ZFCNr",
				Title: "Instagram BlIka1ZFCNr",
				Size:  3003662,
			},
		},
		{
			name: "image test",
			args: test.Args{
				URL:   "https://www.instagram.com/p/Bl5oVUyl9Yx",
				Title: "Instagram Bl5oVUyl9Yx",
				Size:  250596,
			},
		},
		{
			name: "image album test",
			args: test.Args{
				URL:   "https://www.instagram.com/p/Bjyr-gxF4Rb",
				Title: "Instagram Bjyr-gxF4Rb",
				Size:  4599909,
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
