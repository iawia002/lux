package ixigua

type xiguanData struct {
	AnyVideo struct {
		GidInformation struct {
			Gid        string `json:"gid"`
			PackerData struct {
				Video struct {
					Title         string `json:"title"`
					PosterUrl     string `json:"poster_url"`
					VideoResource struct {
						Vid    string `json:"vid"`
						Normal struct {
							VideoId   string `json:"video_id"`
							VideoList map[string]struct {
								Definition  string `json:"definition"`
								Quality     string `json:"quality"`
								Vtype       string `json:"vtype"`
								Vwidth      int    `json:"vwidth"`
								Vheight     int    `json:"vheight"`
								Bitrate     int64  `json:"bitrate"`
								RealBitrate int64  `json:"real_bitrate"`
								Fps         int    `json:"fps"`
								CodecType   string `json:"codec_type"`
								Size        int64  `json:"size"`
								MainUrl     string `json:"main_url"`
								BackupUrl1  string `json:"backup_url_1"`
							} `json:"video_list"`
						} `json:"normal"`
					} `json:"videoResource"`
				} `json:"video"`
				Key string `json:"key"`
			} `json:"packerData"`
		} `json:"gidInformation"`
	} `json:"anyVideo"`
}
