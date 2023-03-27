package zhihu

// minimum field
type video struct {
	PlayList struct {
		FHD resolution `json:"FHD"`
		HD  resolution `json:"HD"`
		SD  resolution `json:"SD"`
	} `json:"playlist_v2"`
}

// minimum field
type resolution struct {
	Size    int64  `json:"size"`
	Format  string `json:"format"`
	PlayURL string `json:"play_url"`
}
