package proto

type Data struct {
	Duration float64  `json:"duration"`
	Streams  []Stream `json:"streams"`
	Title    string   `json:"title"`
}

type Stream struct {
	Referer   string `json:"referer"`
	Useragent string `json:"useragent"`
	Segs      []Seg  `json:"segs"`
	Type      string `json:"type"`
	Quality   string `json:"quality"`
}

type Seg struct {
	Duration  float64 `json:"duration"`
	Size      int     `json:"size"`
	BackupURL string  `json:"backupUrl"`
	URL       string  `json:"url"`
}

// Extractor implements video data extraction related operations.
type Extractor interface {
	// Extract is the main function to extract the data.
	Extract(url string) (TransformData, error)
}

type TransformData interface {
	// Extract is the main function to extract the data.
	TransformData(url string, quality string) *Data
}
