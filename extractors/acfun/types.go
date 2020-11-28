package acfun

type episodeData struct {
	ItemID      int64  `json:"itemId"`
	EpisodeName string `json:"episodeName"`
	BangumiID   int64  `json:"bangumiId"`
	VideoID     int64  `json:"videoId"`
}

type bangumiData struct {
	episodeData
	BangumiTitle     string `json:"bangumiTitle"`
	CurrentVideoInfo struct {
		KsPlayJSON string `json:"ksPlayJson"`
	} `json:"currentVideoInfo"`
}

type videoInfo struct {
	AdaptationSet []struct {
		Streams streams `json:"representation"`
	} `json:"adaptationSet"`
}

type streams []stream

type episodeList struct {
	Episodes []*episodeData `json:"items"`
}

type stream struct {
	ID           int64  `json:"id"`
	BackURL      string `json:"backUrl"`
	Codecs       string `json:"codecs"`
	URL          string `json:"url"`
	BitRate      int64  `json:"avgBitrate"`
	QualityType  string `json:"qualityType"`
	QualityLabel string `json:"qualityLabel"`
}
