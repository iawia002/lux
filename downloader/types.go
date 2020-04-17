package downloader

// Aria2RPCData defines the data structure of json RPC 2.0 info for Aria2
type Aria2RPCData struct {
	// More info about RPC interface please refer to
	// https://aria2.github.io/manual/en/html/aria2c.html#rpc-interface
	JSONRPC string `json:"jsonrpc"`
	ID      string `json:"id"`
	// For a simple download, only inplemented `addUri`
	Method string `json:"method"`
	// secret, uris, options
	Params [3]interface{} `json:"params"`
}

// Aria2Input is options for `aria2.addUri`
// https://aria2.github.io/manual/en/html/aria2c.html#id3
type Aria2Input struct {
	// The file name of the downloaded file
	Out string `json:"out"`
	// For a simple download, only add headers
	Header []string `json:"header"`
}

// FilePartMeta defines the data structure of file meta info.
type FilePartMeta struct {
	Index float32
	Start int64
	End   int64
	Cur   int64
}
