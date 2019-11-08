package bilibili

type qualityInfo struct {
	Description []string `json:"accept_description"`
	Quality     []int    `json:"accept_quality"`
}

type dURLData struct {
	Size  int64  `json:"size"`
	URL   string `json:"url"`
	Order int    `json:"order"`
}

type bilibiliData struct {
	DURL    []dURLData `json:"durl"`
	Format  string     `json:"format"`
	Quality int        `json:"quality"`
}

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
	Aid  int `json:"aid"`
	Cid  int `json:"cid"`
	ID   int `json:"id"`
	EpID int `json:"ep_id"`
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
	VideoData multiPageVideoData `json:"videoData"`
}

var qualityString = map[int]string{
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
