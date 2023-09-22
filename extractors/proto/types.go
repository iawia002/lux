package proto

// Part is the data structure for a single part of the video stream information.
type Part struct {
	URL       string `json:"url"`
	BackupURL string `json:"backupUrl"`
	Size      int64  `json:"size"`
	Ext       string `json:"ext"`
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
	Segs []*Part `json:"segs"`
	// total size of all urls
	Size int64 `json:"size"`
	// the file extension after video parts merged
	Ext string `json:"ext"`
}

// DataType indicates the type of extracted data, eg: video or image.
type DataType string

const (
	// DataTypeVideo indicates the type of extracted data is the video.
	DataTypeVideo DataType = "video"
	// DataTypeImage indicates the type of extracted data is the image.
	DataTypeImage DataType = "image"
	// DataTypeAudio indicates the type of extracted data is the audio.
	DataTypeAudio DataType = "audio"
)

// Data is the main data structure for the whole video data.
type Data struct {
	// URL is used to record the address of this download
	URL   string   `json:"url"`
	Site  string   `json:"site"`
	Title string   `json:"title"`
	Type  DataType `json:"type"`
	// each stream has it's own Parts and Quality
	Streams []*Stream `json:"streams"`
	// Err is used to record whether an error occurred when extracting the list data
	Err error `json:"err"`
}

// FillUpStreamsData fills up some data automatically.
func (d *Data) FillUpStreamsData() {
	for _, stream := range d.Streams {
		if stream.Quality == "" {
			stream.Quality = stream.ID
		}
		// generate the merged file extension
		if d.Type == DataTypeVideo && stream.Ext == "" {
			ext := stream.Segs[0].Ext
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
		for _, part := range stream.Segs {
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

// Extractor implements video data extraction related operations.
type Extractor interface {
	// Extract is the main function to extract the data.
	Extract(url string) (*Data, error)
}
