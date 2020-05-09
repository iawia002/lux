package test

import (
	"sort"
	"testing"

	"github.com/iawia002/annie/extractors/types"
)

// Args Arguments for extractor tests
type Args struct {
	URL     string
	Title   string
	Quality string
	Size    int64
}

// CheckData check the given data
func CheckData(args, data Args) bool {
	if args.Title != data.Title {
		return false
	}
	// not every video got quality information
	if args.Quality != "" && args.Quality != data.Quality {
		return false
	}
	if args.Size != 0 && args.Size != data.Size {
		return false
	}
	return true
}

// Check check the result
func Check(t *testing.T, args Args, data *types.Data) {
	// get the default stream
	sortedStreams := make([]*types.Stream, 0, len(data.Streams))
	for _, s := range data.Streams {
		sortedStreams = append(sortedStreams, s)
	}
	sort.Slice(sortedStreams, func(i, j int) bool { return sortedStreams[i].Size > sortedStreams[j].Size })
	defaultData := sortedStreams[0]

	temp := Args{
		Title:   data.Title,
		Quality: defaultData.Quality,
		Size:    defaultData.Size,
	}
	if !CheckData(args, temp) {
		t.Errorf("Got: %v\nExpected: %v", temp, args)
	}
}

// CheckError check the error
func CheckError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}
