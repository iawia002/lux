package reddit

import (
	"fmt"
	"log"
	"testing"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/test"
)

func TestReddit(t *testing.T) {
	tests := []struct {
		name     string
		args     test.Args
		playlist bool
	}{
		{
			name: "normal test 1",
			args: test.Args{
				URL:   "https://www.reddit.com/r/space/comments/uj8sod/a_couple_of_days_ago_i_visited_this_place_an/",
				Title: "A couple of days ago I visited this place. An abandoned space shuttle",
			},
			playlist: false,
		},
		{
			name: "normal test 2",
			args: test.Args{
				URL:   "https://www.reddit.com/r/DotA2/comments/uq012r/til_how_useful_hurricane_bird_is/",
				Title: "TIL how useful hurricane bird is",
			},
			playlist: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				data []*extractors.Data
				err  error
			)
			data, err = New().Extract(tt.args.URL, extractors.Options{})
			fmt.Print(data)
			if err != nil {
				log.Fatal(err)
			}
			// test.CheckError(t, err)
			// test.Check(t, tt.args, data[0])
		})
	}
}
