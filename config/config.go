package config

var (
	// Debug debug mode
	Debug bool
	// Version show version
	Version bool
	// InfoOnly Information only mode
	InfoOnly bool
	// Cookie http cookies
	Cookie string
	// Playlist download playlist
	Playlist bool
	// Refer use specified Referrer
	Refer string
	// Proxy HTTP proxy
	Proxy string
	// Socks5Proxy SOCKS5 proxy
	Socks5Proxy string
	// Format select specified format to download
	Format string
	// OutputPath output file path
	OutputPath string
	// OutputName output file name
	OutputName string
	// ExtractedData print extracted data
	ExtractedData bool
	// ThreadNumber The number of download thread (only works for multiple-parts video)
	ThreadNumber int
	// File URLs file path
	File string
	// PlaylistStart Playlist video to start at
	PlaylistStart int
	// PlaylistEnd Playlist video to end at
	PlaylistEnd int
	// PlaylistItems Playlist video items to download. Separated by commas like: 1,5,6
	PlaylistItems string
	// Caption download captions
	Caption bool
	// Ccode youku ccode
	Ccode string
)

// FakeHeaders fake http headers
var FakeHeaders = map[string]string{
	"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
	"Accept-Charset":  "UTF-8,*;q=0.5",
	"Accept-Encoding": "gzip,deflate,sdch",
	"Accept-Language": "en-US,en;q=0.8",
	"User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.146 Safari/537.36",
}
