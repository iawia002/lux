package bilibili

type bangumiEpData struct {
	Aid         int    `json:"aid"`
	Cid         int    `json:"cid"`
	BVid        string `json:"bvid"`
	ID          int    `json:"id"`
	EpID        int    `json:"ep_id"`
	TitleFormat string `json:"titleFormat"`
	LongTitle   string `json:"long_title"`
}

type bangumiData struct {
	EpInfo bangumiEpData   `json:"epInfo"`
	EpList []bangumiEpData `json:"epList"`
}

type videoPagesData struct {
	Cid  int    `json:"cid"`
	Part string `json:"part"`
	Page int    `json:"page"`
}

type multiPageVideoData struct {
	Title string           `json:"title"`
	Pages []videoPagesData `json:"pages"`
}

type episode struct {
	Aid   int    `json:"aid"`
	Cid   int    `json:"cid"`
	Title string `json:"title"`
	BVid  string `json:"bvid"`
}

type multiEpisodeData struct {
	Seasionid int       `json:"season_id"`
	Episodes  []episode `json:"episodes"`
}

type multiPage struct {
	Aid       int                `json:"aid"`
	BVid      string             `json:"bvid"`
	Sections  []multiEpisodeData `json:"sections"`
	VideoData multiPageVideoData `json:"videoData"`
}

type dashStream struct {
	ID        int    `json:"id"`
	BaseURL   string `json:"baseUrl"`
	Bandwidth int    `json:"bandwidth"`
	MimeType  string `json:"mimeType"`
	Codecid   int    `json:"codecid"`
	Codecs    string `json:"codecs"`
}

type dashStreams struct {
	Video []dashStream `json:"video"`
	Audio []dashStream `json:"audio"`
}

type dashInfo struct {
	CurQuality  int         `json:"quality"`
	Description []string    `json:"accept_description"`
	Quality     []int       `json:"accept_quality"`
	Streams     dashStreams `json:"dash"`
	DURLFormat  string      `json:"format"`
	DURLs       []dURL      `json:"durl"`
}

type dURL struct {
	URL  string `json:"url"`
	Size int64  `json:"size"`
}

type dash struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    dashInfo `json:"data"`
	Result  dashInfo `json:"result"`
}

var qualityString = map[int]string{
	127: "超高清 8K",
	120: "超清 4K",
	116: "高清 1080P60",
	74:  "高清 720P60",
	112: "高清 1080P+",
	80:  "高清 1080P",
	64:  "高清 720P",
	48:  "高清 720P",
	32:  "清晰 480P",
	16:  "流畅 360P",
	15:  "流畅 360P",
}
