package types

// Part is the data structure for a single part of the video stream information.
type Part struct {
	URL  string `json:"url"`
	Size int64  `json:"size"`
	Ext  string `json:"ext"`
}

// Stream is the data structure for each video stream, eg: 720P, 1080P.
type Stream struct {
	// eg: "1080"
	ID string `json:"id"`
	// eg: "1080P xxx"
	Quality string `json:"quality"`
	// [Part: {URL, Size, Ext}, ...]
	// Some video stream have multiple parts,
	// and can also be used to download multiple image files at once
	Parts []*Part `json:"parts"`
	// total size of all urls
	Size int64 `json:"size"`
	// the file extension after video parts merged
	Ext string `json:"ext"`
	// if the parts need mux
	NeedMux bool
}

// DataType indicates the type of extracted data, eg: video or image.
type DataType string

const (
	// DataTypeVideo indicates the type of extracted data is the video.
	DataTypeVideo DataType = "video"
	// DataTypeImage indicates the type of extracted data is the image.
	DataTypeImage DataType = "image"
)

// Data is the main data structure for the whole video data.
type Data struct {
	// URL is used to record the address of this download
	URL   string   `json:"url"`
	Site  string   `json:"site"`
	Title string   `json:"title"`
	Type  DataType `json:"type"`
	// each stream has it's own Parts and Quality
	Streams map[string]*Stream `json:"streams"`
	// danmaku, subtitles, etc
	Caption *Part `json:"caption"`
	// Err is used to record whether an error occurred when extracting the list data
	Err error `json:"err"`
}

// FillUpStreamsData fills up some data automatically.
func (d *Data) FillUpStreamsData() {
	for id, stream := range d.Streams {
		// fill up ID
		stream.ID = id
		if stream.Quality == "" {
			stream.Quality = id
		}

		// generate the merged file extension
		if d.Type == DataTypeVideo && stream.Ext == "" {
			ext := stream.Parts[0].Ext
			// The file extension in `Parts` is used as the merged file extension by default, except for the following formats
			switch ext {
			// ts and flv files should be merged into an mp4 file
			case "ts", "flv", "f4v":
				ext = "mp4"
			}
			stream.Ext = ext
		}

		// calculate total size
		if stream.Size > 0 {
			continue
		}
		var size int64
		for _, part := range stream.Parts {
			size += part.Size
		}
		stream.Size = size
	}
}

// EmptyData returns an "empty" Data object with the given URL and error.
func EmptyData(url string, err error) *Data {
	return &Data{
		URL: url,
		Err: err,
	}
}

// Options defines optional options that can be used in the extraction function.
type Options struct {
	// Playlist indicates if we need to extract the whole playlist rather than the single video.
	Playlist bool
	// Items defines wanted items from a playlist. Separated by commas like: 1,5,6,8-10.
	Items string
	// ItemStart defines the starting item of a playlist.
	ItemStart int
	// ItemEnd defines the ending item of a playlist.
	ItemEnd int

	// ThreadNumber defines how many threads will use in the extraction, only works when Playlist is true.
	ThreadNumber int
	Cookie       string

	// EpisodeTitleOnly indicates file name of each bilibili episode doesn't include the playlist title
	EpisodeTitleOnly bool

	YoukuCcode    string
	YoukuCkey     string
	YoukuPassword string
}

// Extractor implements video data extraction related operations.
type Extractor interface {
	// Extract is the main function to extract the data.
	Extract(url string, option Options) ([]*Data, error)
}
