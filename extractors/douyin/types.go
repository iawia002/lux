package douyin

type douyinData struct {
	StatusCode  int `json:"status_code"`
	AwemeDetail struct {
		AdmireAuth struct {
			AdmireButton       int `json:"admire_button"`
			IsAdmire           int `json:"is_admire"`
			IsShowAdmireButton int `json:"is_show_admire_button"`
			IsShowAdmireTab    int `json:"is_show_admire_tab"`
		} `json:"admire_auth"`
		Anchors interface{} `json:"anchors"`
		Author  struct {
			AvatarThumb struct {
				Height  int      `json:"height"`
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
				Width   int      `json:"width"`
			} `json:"avatar_thumb"`
			CfList          interface{} `json:"cf_list"`
			CloseFriendType int         `json:"close_friend_type"`
			ContactsStatus  int         `json:"contacts_status"`
			ContrailList    interface{} `json:"contrail_list"`
			CoverURL        []struct {
				Height  int      `json:"height"`
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
				Width   int      `json:"width"`
			} `json:"cover_url"`
			CreateTime                             int         `json:"create_time"`
			CustomVerify                           string      `json:"custom_verify"`
			DataLabelList                          interface{} `json:"data_label_list"`
			EndorsementInfoList                    interface{} `json:"endorsement_info_list"`
			EnterpriseVerifyReason                 string      `json:"enterprise_verify_reason"`
			FavoritingCount                        int         `json:"favoriting_count"`
			FollowStatus                           int         `json:"follow_status"`
			FollowerCount                          int         `json:"follower_count"`
			FollowerListSecondaryInformationStruct interface{} `json:"follower_list_secondary_information_struct"`
			FollowerStatus                         int         `json:"follower_status"`
			FollowingCount                         int         `json:"following_count"`
			ImRoleIds                              interface{} `json:"im_role_ids"`
			IsAdFake                               bool        `json:"is_ad_fake"`
			IsBlockedV2                            bool        `json:"is_blocked_v2"`
			IsBlockingV2                           bool        `json:"is_blocking_v2"`
			IsCf                                   int         `json:"is_cf"`
			MaxFollowerCount                       int         `json:"max_follower_count"`
			Nickname                               string      `json:"nickname"`
			NotSeenItemIDList                      interface{} `json:"not_seen_item_id_list"`
			NotSeenItemIDListV2                    interface{} `json:"not_seen_item_id_list_v2"`
			OfflineInfoList                        interface{} `json:"offline_info_list"`
			PersonalTagList                        interface{} `json:"personal_tag_list"`
			PreventDownload                        bool        `json:"prevent_download"`
			RiskNoticeText                         string      `json:"risk_notice_text"`
			SecUID                                 string      `json:"sec_uid"`
			Secret                                 int         `json:"secret"`
			ShareInfo                              struct {
				ShareDesc      string `json:"share_desc"`
				ShareDescInfo  string `json:"share_desc_info"`
				ShareQrcodeURL struct {
					Height  int      `json:"height"`
					URI     string   `json:"uri"`
					URLList []string `json:"url_list"`
					Width   int      `json:"width"`
				} `json:"share_qrcode_url"`
				ShareTitle       string `json:"share_title"`
				ShareTitleMyself string `json:"share_title_myself"`
				ShareTitleOther  string `json:"share_title_other"`
				ShareURL         string `json:"share_url"`
				ShareWeiboDesc   string `json:"share_weibo_desc"`
			} `json:"share_info"`
			ShortID             string      `json:"short_id"`
			Signature           string      `json:"signature"`
			SignatureExtra      interface{} `json:"signature_extra"`
			SpecialPeopleLabels interface{} `json:"special_people_labels"`
			Status              int         `json:"status"`
			TextExtra           interface{} `json:"text_extra"`
			TotalFavorited      int         `json:"total_favorited"`
			UID                 string      `json:"uid"`
			UniqueID            string      `json:"unique_id"`
			UserAge             int         `json:"user_age"`
			UserCanceled        bool        `json:"user_canceled"`
			UserPermissions     interface{} `json:"user_permissions"`
			VerificationType    int         `json:"verification_type"`
		} `json:"author"`
		AuthorMaskTag int   `json:"author_mask_tag"`
		AuthorUserID  int64 `json:"author_user_id"`
		AwemeACL      struct {
			DownloadMaskPanel struct {
				Code     int `json:"code"`
				ShowType int `json:"show_type"`
			} `json:"download_mask_panel"`
		} `json:"aweme_acl"`
		AwemeControl struct {
			CanComment     bool `json:"can_comment"`
			CanForward     bool `json:"can_forward"`
			CanShare       bool `json:"can_share"`
			CanShowComment bool `json:"can_show_comment"`
		} `json:"aweme_control"`
		AwemeID               string      `json:"aweme_id"`
		AwemeType             int         `json:"aweme_type"`
		ChallengePosition     interface{} `json:"challenge_position"`
		ChapterList           interface{} `json:"chapter_list"`
		CollectStat           int         `json:"collect_stat"`
		CommentGid            int64       `json:"comment_gid"`
		CommentList           interface{} `json:"comment_list"`
		CommentPermissionInfo struct {
			CanComment              bool `json:"can_comment"`
			CommentPermissionStatus int  `json:"comment_permission_status"`
			ItemDetailEntry         bool `json:"item_detail_entry"`
			PressEntry              bool `json:"press_entry"`
			ToastGuide              bool `json:"toast_guide"`
		} `json:"comment_permission_info"`
		CommerceConfigData interface{} `json:"commerce_config_data"`
		CommonBarInfo      string      `json:"common_bar_info"`
		ComponentInfoV2    string      `json:"component_info_v2"`
		CoverLabels        interface{} `json:"cover_labels"`
		CreateTime         int         `json:"create_time"`
		Desc               string      `json:"desc"`
		DiggLottie         struct {
			CanBomb  int    `json:"can_bomb"`
			LottieID string `json:"lottie_id"`
		} `json:"digg_lottie"`
		DisableRelationBar      int         `json:"disable_relation_bar"`
		DislikeDimensionList    interface{} `json:"dislike_dimension_list"`
		DuetAggregateInMusicTab bool        `json:"duet_aggregate_in_music_tab"`
		Duration                int         `json:"duration"`
		FeedCommentConfig       struct {
			AuthorAuditStatus int    `json:"author_audit_status"`
			InputConfigText   string `json:"input_config_text"`
		} `json:"feed_comment_config"`
		Geofencing          []interface{} `json:"geofencing"`
		GeofencingRegions   interface{}   `json:"geofencing_regions"`
		GroupID             string        `json:"group_id"`
		HybridLabel         interface{}   `json:"hybrid_label"`
		ImageAlbumMusicInfo struct {
			BeginTime int `json:"begin_time"`
			EndTime   int `json:"end_time"`
			Volume    int `json:"volume"`
		} `json:"image_album_music_info"`
		ImageInfos interface{} `json:"image_infos"`
		ImageList  interface{} `json:"image_list"`
		Images     []struct {
			DownloadURLList []string    `json:"download_url_list"`
			Height          int         `json:"height"`
			MaskURLList     interface{} `json:"mask_url_list"`
			URI             string      `json:"uri"`
			URLList         []string    `json:"url_list"`
			Width           int         `json:"width"`
		} `json:"images"`
		ImgBitrate []struct {
			Images []struct {
				DownloadURLList []string    `json:"download_url_list"`
				Height          int         `json:"height"`
				MaskURLList     interface{} `json:"mask_url_list"`
				URI             string      `json:"uri"`
				URLList         []string    `json:"url_list"`
				Width           int         `json:"width"`
			} `json:"images"`
			Name string `json:"name"`
		} `json:"img_bitrate"`
		ImpressionData struct {
			GroupIDListA   []int64     `json:"group_id_list_a"`
			GroupIDListB   []int64     `json:"group_id_list_b"`
			SimilarIDListA interface{} `json:"similar_id_list_a"`
			SimilarIDListB interface{} `json:"similar_id_list_b"`
		} `json:"impression_data"`
		InteractionStickers  interface{} `json:"interaction_stickers"`
		IsAds                bool        `json:"is_ads"`
		IsCollectsSelected   int         `json:"is_collects_selected"`
		IsDuetSing           bool        `json:"is_duet_sing"`
		IsImageBeat          bool        `json:"is_image_beat"`
		IsLifeItem           bool        `json:"is_life_item"`
		IsMultiContent       int         `json:"is_multi_content"`
		IsStory              int         `json:"is_story"`
		IsTop                int         `json:"is_top"`
		ItemWarnNotification struct {
			Content string `json:"content"`
			Show    bool   `json:"show"`
			Type    int    `json:"type"`
		} `json:"item_warn_notification"`
		LabelTopText interface{} `json:"label_top_text"`
		LongVideo    interface{} `json:"long_video"`
		Music        struct {
			Album            string        `json:"album"`
			ArtistUserInfos  interface{}   `json:"artist_user_infos"`
			Artists          []interface{} `json:"artists"`
			AuditionDuration int           `json:"audition_duration"`
			Author           string        `json:"author"`
			AuthorDeleted    bool          `json:"author_deleted"`
			AuthorPosition   interface{}   `json:"author_position"`
			AuthorStatus     int           `json:"author_status"`
			AvatarLarge      struct {
				Height  int      `json:"height"`
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
				Width   int      `json:"width"`
			} `json:"avatar_large"`
			AvatarMedium struct {
				Height  int      `json:"height"`
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
				Width   int      `json:"width"`
			} `json:"avatar_medium"`
			AvatarThumb struct {
				Height  int      `json:"height"`
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
				Width   int      `json:"width"`
			} `json:"avatar_thumb"`
			BindedChallengeID int  `json:"binded_challenge_id"`
			CanBackgroundPlay bool `json:"can_background_play"`
			CollectStat       int  `json:"collect_stat"`
			CoverColorHsv     struct {
				H int `json:"h"`
				S int `json:"s"`
				V int `json:"v"`
			} `json:"cover_color_hsv"`
			CoverHd struct {
				Height  int      `json:"height"`
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
				Width   int      `json:"width"`
			} `json:"cover_hd"`
			CoverLarge struct {
				Height  int      `json:"height"`
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
				Width   int      `json:"width"`
			} `json:"cover_large"`
			CoverMedium struct {
				Height  int      `json:"height"`
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
				Width   int      `json:"width"`
			} `json:"cover_medium"`
			CoverThumb struct {
				Height  int      `json:"height"`
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
				Width   int      `json:"width"`
			} `json:"cover_thumb"`
			DmvAutoShow          bool          `json:"dmv_auto_show"`
			DspStatus            int           `json:"dsp_status"`
			Duration             int           `json:"duration"`
			EndTime              int           `json:"end_time"`
			ExternalSongInfo     []interface{} `json:"external_song_info"`
			Extra                string        `json:"extra"`
			ID                   int64         `json:"id"`
			IDStr                string        `json:"id_str"`
			IsAudioURLWithCookie bool          `json:"is_audio_url_with_cookie"`
			IsCommerceMusic      bool          `json:"is_commerce_music"`
			IsDelVideo           bool          `json:"is_del_video"`
			IsMatchedMetadata    bool          `json:"is_matched_metadata"`
			IsOriginal           bool          `json:"is_original"`
			IsOriginalSound      bool          `json:"is_original_sound"`
			IsPgc                bool          `json:"is_pgc"`
			IsRestricted         bool          `json:"is_restricted"`
			IsVideoSelfSee       bool          `json:"is_video_self_see"`
			LunaInfo             struct {
				HasCopyright bool `json:"has_copyright"`
				IsLunaUser   bool `json:"is_luna_user"`
			} `json:"luna_info"`
			LyricShortPosition interface{} `json:"lyric_short_position"`
			MatchedPgcSound    struct {
				Author      string `json:"author"`
				CoverMedium struct {
					Height  int      `json:"height"`
					URI     string   `json:"uri"`
					URLList []string `json:"url_list"`
					Width   int      `json:"width"`
				} `json:"cover_medium"`
				MixedAuthor string `json:"mixed_author"`
				MixedTitle  string `json:"mixed_title"`
				Title       string `json:"title"`
			} `json:"matched_pgc_sound"`
			Mid               string      `json:"mid"`
			MusicChartRanks   interface{} `json:"music_chart_ranks"`
			MusicStatus       int         `json:"music_status"`
			MusicianUserInfos interface{} `json:"musician_user_infos"`
			MuteShare         bool        `json:"mute_share"`
			OfflineDesc       string      `json:"offline_desc"`
			OwnerHandle       string      `json:"owner_handle"`
			OwnerID           string      `json:"owner_id"`
			OwnerNickname     string      `json:"owner_nickname"`
			PgcMusicType      int         `json:"pgc_music_type"`
			PlayURL           struct {
				Height  int      `json:"height"`
				URI     string   `json:"uri"`
				URLKey  string   `json:"url_key"`
				URLList []string `json:"url_list"`
				Width   int      `json:"width"`
			} `json:"play_url"`
			Position                  interface{} `json:"position"`
			PreventDownload           bool        `json:"prevent_download"`
			PreventItemDownloadStatus int         `json:"prevent_item_download_status"`
			PreviewEndTime            int         `json:"preview_end_time"`
			PreviewStartTime          float64     `json:"preview_start_time"`
			ReasonType                int         `json:"reason_type"`
			Redirect                  bool        `json:"redirect"`
			SchemaURL                 string      `json:"schema_url"`
			SearchImpr                struct {
				EntityID string `json:"entity_id"`
			} `json:"search_impr"`
			SecUID        string `json:"sec_uid"`
			ShootDuration int    `json:"shoot_duration"`
			Song          struct {
				Artists interface{} `json:"artists"`
				ID      int64       `json:"id"`
				IDStr   string      `json:"id_str"`
			} `json:"song"`
			SourcePlatform    int         `json:"source_platform"`
			StartTime         int         `json:"start_time"`
			Status            int         `json:"status"`
			TagList           interface{} `json:"tag_list"`
			Title             string      `json:"title"`
			UnshelveCountries interface{} `json:"unshelve_countries"`
			UserCount         int         `json:"user_count"`
			VideoDuration     int         `json:"video_duration"`
		} `json:"music"`
		NicknamePosition    interface{}   `json:"nickname_position"`
		OriginCommentIds    interface{}   `json:"origin_comment_ids"`
		OriginTextExtra     []interface{} `json:"origin_text_extra"`
		OriginalImages      interface{}   `json:"original_images"`
		PackedClips         interface{}   `json:"packed_clips"`
		PhotoSearchEntrance struct {
			EcomType int `json:"ecom_type"`
		} `json:"photo_search_entrance"`
		Position           interface{}   `json:"position"`
		PressPanelInfo     string        `json:"press_panel_info"`
		PreviewTitle       string        `json:"preview_title"`
		PreviewVideoStatus int           `json:"preview_video_status"`
		Promotions         []interface{} `json:"promotions"`
		Rate               int           `json:"rate"`
		Region             string        `json:"region"`
		RelationLabels     interface{}   `json:"relation_labels"`
		SearchImpr         struct {
			EntityID   string `json:"entity_id"`
			EntityType string `json:"entity_type"`
		} `json:"search_impr"`
		SeriesPaidInfo struct {
			ItemPrice        int `json:"item_price"`
			SeriesPaidStatus int `json:"series_paid_status"`
		} `json:"series_paid_info"`
		ShareInfo struct {
			ShareDesc     string `json:"share_desc"`
			ShareDescInfo string `json:"share_desc_info"`
			ShareLinkDesc string `json:"share_link_desc"`
			ShareURL      string `json:"share_url"`
		} `json:"share_info"`
		ShareURL           string `json:"share_url"`
		ShouldOpenAdReport bool   `json:"should_open_ad_report"`
		ShowFollowButton   struct {
		} `json:"show_follow_button"`
		SocialTagList       interface{} `json:"social_tag_list"`
		StandardBarInfoList interface{} `json:"standard_bar_info_list"`
		Statistics          struct {
			AdmireCount  int    `json:"admire_count"`
			AwemeID      string `json:"aweme_id"`
			CollectCount int    `json:"collect_count"`
			CommentCount int    `json:"comment_count"`
			DiggCount    int    `json:"digg_count"`
			PlayCount    int    `json:"play_count"`
			ShareCount   int    `json:"share_count"`
		} `json:"statistics"`
		Status struct {
			AllowShare        bool   `json:"allow_share"`
			AwemeID           string `json:"aweme_id"`
			InReviewing       bool   `json:"in_reviewing"`
			IsDelete          bool   `json:"is_delete"`
			IsProhibited      bool   `json:"is_prohibited"`
			ListenVideoStatus int    `json:"listen_video_status"`
			PartSee           int    `json:"part_see"`
			PrivateStatus     int    `json:"private_status"`
			ReviewResult      struct {
				ReviewStatus int `json:"review_status"`
			} `json:"review_result"`
		} `json:"status"`
		TextExtra []struct {
			End         int    `json:"end"`
			HashtagID   string `json:"hashtag_id"`
			HashtagName string `json:"hashtag_name"`
			IsCommerce  bool   `json:"is_commerce"`
			Start       int    `json:"start"`
			Type        int    `json:"type"`
		} `json:"text_extra"`
		UniqidPosition interface{} `json:"uniqid_position"`
		UserDigged     int         `json:"user_digged"`
		Video          struct {
			BigThumbs []struct {
				Duration float64 `json:"duration"`
				Fext     string  `json:"fext"`
				ImgNum   int     `json:"img_num"`
				ImgURL   string  `json:"img_url"`
				ImgXLen  int     `json:"img_x_len"`
				ImgXSize int     `json:"img_x_size"`
				ImgYLen  int     `json:"img_y_len"`
				ImgYSize int     `json:"img_y_size"`
				Interval float64 `json:"interval"`
				URI      string  `json:"uri"`
			} `json:"big_thumbs"`
			BitRate []struct {
				FPS       int    `json:"FPS"`
				HDRBit    string `json:"HDR_bit"`
				HDRType   string `json:"HDR_type"`
				BitRate   int    `json:"bit_rate"`
				GearName  string `json:"gear_name"`
				IsBytevc1 int    `json:"is_bytevc1"`
				IsH265    int    `json:"is_h265"`
				PlayAddr  struct {
					DataSize int      `json:"data_size"`
					FileCs   string   `json:"file_cs"`
					FileHash string   `json:"file_hash"`
					Height   int      `json:"height"`
					URI      string   `json:"uri"`
					URLKey   string   `json:"url_key"`
					URLList  []string `json:"url_list"`
					Width    int      `json:"width"`
				} `json:"play_addr"`
				QualityType int `json:"quality_type"`
			} `json:"bit_rate"`
			Cover struct {
				Height  int      `json:"height"`
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
				Width   int      `json:"width"`
			} `json:"cover"`
			CoverOriginalScale struct {
				Height  int      `json:"height"`
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
				Width   int      `json:"width"`
			} `json:"cover_original_scale"`
			Duration     int `json:"duration"`
			DynamicCover struct {
				Height  int      `json:"height"`
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
				Width   int      `json:"width"`
			} `json:"dynamic_cover"`
			Height      int    `json:"height"`
			IsH265      int    `json:"is_h265"`
			IsLongVideo int    `json:"is_long_video"`
			IsSourceHDR int    `json:"is_source_HDR"`
			Meta        string `json:"meta"`
			OriginCover struct {
				Height  int      `json:"height"`
				URI     string   `json:"uri"`
				URLList []string `json:"url_list"`
				Width   int      `json:"width"`
			} `json:"origin_cover"`
			PlayAddr struct {
				DataSize int      `json:"data_size"`
				FileCs   string   `json:"file_cs"`
				FileHash string   `json:"file_hash"`
				Height   int      `json:"height"`
				URI      string   `json:"uri"`
				URLKey   string   `json:"url_key"`
				URLList  []string `json:"url_list"`
				Width    int      `json:"width"`
			} `json:"play_addr"`
			PlayAddr265 struct {
				DataSize int      `json:"data_size"`
				FileCs   string   `json:"file_cs"`
				FileHash string   `json:"file_hash"`
				Height   int      `json:"height"`
				URI      string   `json:"uri"`
				URLKey   string   `json:"url_key"`
				URLList  []string `json:"url_list"`
				Width    int      `json:"width"`
			} `json:"play_addr_265"`
			PlayAddrH264 struct {
				DataSize int      `json:"data_size"`
				FileCs   string   `json:"file_cs"`
				FileHash string   `json:"file_hash"`
				Height   int      `json:"height"`
				URI      string   `json:"uri"`
				URLKey   string   `json:"url_key"`
				URLList  []string `json:"url_list"`
				Width    int      `json:"width"`
			} `json:"play_addr_h264"`
			Ratio      string `json:"ratio"`
			VideoModel string `json:"video_model"`
			Width      int    `json:"width"`
		} `json:"video"`
		VideoLabels interface{} `json:"video_labels"`
		VideoTag    []struct {
			Level   int    `json:"level"`
			TagID   int    `json:"tag_id"`
			TagName string `json:"tag_name"`
		} `json:"video_tag"`
		VideoText []interface{} `json:"video_text"`
		WannaTag  struct {
		} `json:"wanna_tag"`
	} `json:"aweme_detail"`
	Extra struct {
		Now   int64  `json:"now"`
		Logid string `json:"logid"`
	} `json:"extra"`
}
