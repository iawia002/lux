package test

import (
	"sort"
	"testing"

	"github.com/iawia002/annie/downloader"
)

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
func Check(t *testing.T, args Args, data downloader.VideoData) {
	// get the default format
	sortedFormats := make([]downloader.FormatData, len(data.Formats))
	for _, data := range data.Formats {
		sortedFormats = append(sortedFormats, data)
	}
	sort.Slice(sortedFormats, func(i, j int) bool { return sortedFormats[i].Size > sortedFormats[j].Size })
	defaultData := sortedFormats[0]

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
