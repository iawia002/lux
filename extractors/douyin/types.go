package douyin

type douyinData struct {
	StatusCode int `json:"status_code"`
	ItemList   []struct {
		CommentList interface{}   `json:"comment_list"`
		GroupID     int64         `json:"group_id"`
		TextExtra   []interface{} `json:"text_extra"`
		ImageInfos  interface{}   `json:"image_infos"`
		AwemeID     string        `json:"aweme_id"`
		ShareInfo   struct {
			ShareWeiboDesc string `json:"share_weibo_desc"`
			ShareDesc      string `json:"share_desc"`
			ShareTitle     string `json:"share_title"`
		} `json:"share_info"`
		IsPreview int         `json:"is_preview"`
		Images    interface{} `json:"images"`
		RiskInfos struct {
			Type    int    `json:"type"`
			Content string `json:"content"`
			Warn    bool   `json:"warn"`
		} `json:"risk_infos"`
		VideoText    interface{} `json:"video_text"`
		LabelTopText interface{} `json:"label_top_text"`
		Author       struct {
			Nickname    string `json:"nickname"`
			AvatarThumb struct {
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
			} `json:"avatar_thumb"`
			AvatarMedium struct {
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
			} `json:"avatar_medium"`
			PlatformSyncInfo interface{} `json:"platform_sync_info"`
			PolicyVersion    interface{} `json:"policy_version"`
			UID              string      `json:"uid"`
			ShortID          string      `json:"short_id"`
			Signature        string      `json:"signature"`
			AvatarLarger     struct {
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
			} `json:"avatar_larger"`
			UniqueID        string      `json:"unique_id"`
			FollowersDetail interface{} `json:"followers_detail"`
			Geofencing      interface{} `json:"geofencing"`
			TypeLabel       interface{} `json:"type_label"`
		} `json:"author"`
		Music struct {
			CoverMedium struct {
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
			} `json:"cover_medium"`
			CoverThumb struct {
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
			} `json:"cover_thumb"`
			Duration int    `json:"duration"`
			Status   int    `json:"status"`
			Author   string `json:"author"`
			Mid      string `json:"mid"`
			Title    string `json:"title"`
			CoverHd  struct {
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
			} `json:"cover_hd"`
			CoverLarge struct {
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
			} `json:"cover_large"`
			PlayURL struct {
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
			} `json:"play_url"`
			Position interface{} `json:"position"`
			ID       int64       `json:"id"`
		} `json:"music"`
		ChaList interface{} `json:"cha_list"`
		Video   struct {
			PlayAddr struct {
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
			} `json:"play_addr"`
			Width        int    `json:"width"`
			Ratio        string `json:"ratio"`
			HasWatermark bool   `json:"has_watermark"`
			Duration     int    `json:"duration"`
			Vid          string `json:"vid"`
			Cover        struct {
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
			} `json:"cover"`
			Height       int `json:"height"`
			DynamicCover struct {
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
			} `json:"dynamic_cover"`
			OriginCover struct {
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
			} `json:"origin_cover"`
			BitRate interface{} `json:"bit_rate"`
		} `json:"video"`
		ShareURL     string      `json:"share_url"`
		AuthorUserID int64       `json:"author_user_id"`
		Geofencing   interface{} `json:"geofencing"`
		Promotions   interface{} `json:"promotions"`
		LongVideo    interface{} `json:"long_video"`
		ForwardID    string      `json:"forward_id"`
		Desc         string      `json:"desc"`
		CreateTime   int         `json:"create_time"`
		Statistics   struct {
			CommentCount int    `json:"comment_count"`
			DiggCount    int    `json:"digg_count"`
			PlayCount    int    `json:"play_count"`
			ShareCount   int    `json:"share_count"`
			AwemeID      string `json:"aweme_id"`
		} `json:"statistics"`
		VideoLabels  interface{} `json:"video_labels"`
		Duration     int         `json:"duration"`
		AwemeType    int         `json:"aweme_type"`
		IsLiveReplay bool        `json:"is_live_replay"`
	} `json:"item_list"`
	Extra struct {
		Now   int64  `json:"now"`
		Logid string `json:"logid"`
	} `json:"extra"`
}
