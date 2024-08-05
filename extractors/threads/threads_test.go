package threads_test

import (
	"testing"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/extractors/threads"
	"github.com/iawia002/lux/test"
)

func TestDownload(t *testing.T) {
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "video test",
			args: test.Args{
				URL:   "https://www.threads.net/@rowancheung/post/C9xPmHcpfiN",
				Title: `Threads @rowancheung - C9xPmHcpfiN`,
				Size:  5740684,
			},
		},
		{
			name: "video shared test",
			args: test.Args{
				URL:   "https://www.threads.net/@zuck/post/C9xRqbNPbx2",
				Title: `Threads @zuck - C9xRqbNPbx2`,
				Size:  5740684,
			},
		},
		{
			name: "image test",
			args: test.Args{
				URL:   "https://www.threads.net/@zuck/post/C-BoS7lM8sH",
				Title: `Threads @zuck - C-BoS7lM8sH`,
				Size:  159331,
			},
		},
		{
			name: "hybrid album test",
			args: test.Args{
				URL:   "https://www.threads.net/@meta/post/C95Z1DrPNhi",
				Title: `Threads @meta - C95Z1DrPNhi`,
				Size:  1131229,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := threads.New().Extract(tt.args.URL, extractors.Options{})
			test.CheckError(t, err)
			test.Check(t, tt.args, data[0])
		})
	}
}
