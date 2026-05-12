package extractors

import "testing"

type nilEntryExtractor struct{}

func (nilEntryExtractor) Extract(_ string, _ Options) ([]*Data, error) {
	return []*Data{nil, {URL: "https://luxtest.example.com", Streams: map[string]*Stream{}}}, nil
}

func TestExtractSkipsNilEntries(t *testing.T) {
	Register("luxtest", nilEntryExtractor{})

	data, err := Extract("https://luxtest.com/video", Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("want 1 entry, got %d", len(data))
	}
	if data[0] == nil {
		t.Fatal("returned entry is nil")
	}
}
