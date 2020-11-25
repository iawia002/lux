package bilibili

// {"code":0,"message":"0","ttl":1,"data":{"token":"aaa"}}
// {"code":-101,"message":"账号未登录","ttl":1}
type tokenData struct {
	Token string `json:"token"`
}

type token struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Data    tokenData `json:"data"`
}

type bangumiEpData struct {
	Aid  int    `json:"aid"`
	Cid  int    `json:"cid"`
	BVid string `json:"bvid"`
	ID   int    `json:"id"`
	EpID int    `json:"ep_id"`
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

type multiPage struct {
	Aid       int                `json:"aid"`
	BVid      string             `json:"bvid"`
	VideoData multiPageVideoData `json:"videoData"`
}

type dashStream struct {
	ID        int    `json:"id"`
	BaseURL   string `json:"baseUrl"`
	Bandwidth int    `json:"bandwidth"`
}

type dashStreams struct {
	Video []dashStream `json:"video"`
	Audio []dashStream `json:"audio"`
}

type dURL struct {
	Size int64  `json:"size"`
	URL  string `json:"url"`
}

type dashInfo struct {
	CurQuality  int         `json:"quality"`
	Description []string    `json:"accept_description"`
	Quality     []int       `json:"accept_quality"`
	Streams     dashStreams `json:"dash"`
	DURL        []dURL      `json:"durl"`
}

type dash struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    dashInfo `json:"data"`
	Result  dashInfo `json:"result"`
}

var qualityString = map[int]string{
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
