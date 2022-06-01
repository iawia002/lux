package reddit

import (
	"testing"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/test"
)

func TestReddit(t *testing.T) {
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "normal test 0",
			args: test.Args{
				URL:   "https://www.reddit.com/r/space/comments/uj8sod/a_couple_of_days_ago_i_visited_this_place_an/",
				Title: "A couple of days ago I visited this place. An abandoned space shuttle : space",
			},
		},
		{
			name: "normal test 1",
			args: test.Args{
				URL:   "https://www.reddit.com/r/DotA2/comments/uq012r/til_how_useful_hurricane_bird_is/",
				Title: "TIL how useful hurricane bird is : DotA2",
			},
		},
		{
			name: "normal test 2",
			args: test.Args{
				URL:   "https://www.reddit.com/r/ProgrammerHumor/comments/uqovco/my_code_works/",
				Title: "My code works : ProgrammerHumor",
			},
		},
		{
			name: "normal test 3",
			args: test.Args{
				URL:   "https://www.reddit.com/r/AnimatedPixelArt/comments/uomu32/animation_for_astral_ascent/",
				Title: "Animation for Astral Ascent : AnimatedPixelArt",
			},
		},
		{
			name: "normal test 4",
			args: test.Args{
				URL:   "https://www.reddit.com/r/linuxmemes/comments/v1a4wh/please_olive_do_something/",
				Title: "Please Olive, do something... : linuxmemes",
			},
		},
		{
			name: "normal test 5",
			args: test.Args{
				URL:   "https://www.reddit.com/r/gaming/comments/v27m79/skyrim_probably/",
				Title: "Skyrim, probably : gaming",
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
