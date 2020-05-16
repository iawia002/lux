package pornhub

import (
	"testing"

	"github.com/iawia002/annie/extractors/types"
	"github.com/iawia002/annie/test"
)

func TestPornhub(t *testing.T) {
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "normal test",
			args: test.Args{
				URL:   "https://www.pornhub.com/view_video.php?viewkey=ph5cb5fc41c6ebd",
				Title: "Must watch Milf drilled by the fireplace",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			New().Extract(tt.args.URL, types.Options{})
		})
	}
}
