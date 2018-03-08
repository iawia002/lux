package test

import (
	"testing"

	"github.com/iawia002/annie/downloader"
)

// CheckData check the given data
func CheckData(args Args, data downloader.VideoData) bool {
	if args.Title == data.Title {
		// not every video got quality information
		if args.Quality != "" {
			if args.Quality == data.Quality {
				return true
			}
			return false
		}
		return true
	}
	return false
}

// Check check the result
func Check(t *testing.T, args Args, data downloader.VideoData) {
	if !CheckData(args, data) {
		t.Errorf("Got: %v\nExpected: %v", data, args)
	}
}
