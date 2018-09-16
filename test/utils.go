package test

import (
	"sort"
	"testing"

	"github.com/iawia002/annie/downloader"
)

// CheckData check the given data
func CheckData(args Args, data downloader.VideoData) bool {
	sortedFormats := make([]downloader.FormatData, len(data.Formats))
	for _, data := range data.Formats {
		sortedFormats = append(sortedFormats, data)
	}
	sort.Slice(sortedFormats, func(i, j int) bool { return sortedFormats[i].Size > sortedFormats[j].Size })
	defaultData := sortedFormats[0]
	if args.Title != data.Title {
		return false
	}
	// not every video got quality information
	if args.Quality != "" {
		if args.Quality != defaultData.Quality {
			return false
		}
	}
	if args.Size != 0 {
		if args.Size != defaultData.Size {
			return false
		}
	}
	return true
}

// Check check the result
func Check(t *testing.T, args Args, data downloader.VideoData) {
	if !CheckData(args, data) {
		t.Errorf("Got: %v\nExpected: %v", data, args)
	}
}
