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

type subtitleData struct {
	From     float32 `json:"from"`
	To       float32 `json:"to"`
	Location int     `json:"location"`
	Content  string  `json:"content"`
}

type bilibiliSubtitleFormat struct {
	FontSize        float32        `json:"font_size"`
	FontColor       string         `json:"font_color"`
	BackgroundAlpha float32        `json:"background_alpha"`
	BackgroundColor string         `json:"background_color"`
	Stroke          string         `json:"Stroke"`
	Body            []subtitleData `json:"body"`
}

type subtitleProperty struct {
	ID          int64  `json:"id"`
	Lan         string `json:"lan"`
	LanDoc      string `json:"lan_doc"`
	SubtitleUrl string `json:"subtitle_url"`
}

type subtitleInfo struct {
	AllowSubmit  bool               `json:"allow_submit"`
	SubtitleList []subtitleProperty `json:"list"`
}

type bilibiliWebInterfaceData struct {
	Bvid         string       `json:"bvid"`
	SubtitleInfo subtitleInfo `json:"subtitle"`
}

type bilibiliWebInterface struct {
	Code int                      `json:"code"`
	Data bilibiliWebInterfaceData `json:"data"`
}
