package ximalaya

type ximalayaData struct {
	StatusCode int `json:"ret"`
	Data       struct {
		TrackId         int    `json:"trackId"`
		CanPlay         bool   `json:"canPlay"`
		Src             string `json:"src"`
		XimiVipFreeType int    `json:"ximiVipFreeType"`
		SampleDuration  int    `json:"sampleDuration"`
	} `json:"data"`
}
