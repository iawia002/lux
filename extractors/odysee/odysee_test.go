package odysee

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
			name: "video test 1",
			args: test.Args{
				URL:   "https://odysee.com/@FunnyPets:4e/funny-pets-378-funny-shorts:6",
				Title: "Funny Pets 378 #funny #shorts",
				Size:  1144972,
			},
		},
		{
			name: "video test 2",
			args: test.Args{
				URL:   "https://odysee.com/@FunnyPets:4e/best-of-funny-pets-week-2-funny-pets:a",
				Title: "Best of Funny Pets Week 2 #funny #pets",
				Size:  167272140,
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
