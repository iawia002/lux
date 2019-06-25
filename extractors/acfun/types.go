package acfun

type normalPageInfoData struct {
	Title            string `json:"title"`
	CurrentVideoInfo struct {
		Title     string `json:"title"`
		PlayInfos []struct {
			PlayUrls []string `json:"playUrls"`
		} `json:"playInfos"`
		Id string `json:"id"`
	} `json:"currentVideoInfo"`
}

type bangumiPageInfoData struct {
	Album struct {
		Title string `json:"title"`
	} `json:"album"`
	Video struct {
		Videos []struct {
			NewTitle    string `json:"newTitle"`
			VideoId     int    `json:"videoId"`
			EpisodeName string `json:"episodeName"`
		} `json:"videos"`
	} `json:"video"`
}

type apiInfoData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Embsig   string `json:"embsig"`
		SourceId string `json:"sourceId"`
	}
}

type encryptedInfoData struct {
	E struct {
		Desc string `json:"desc"`
		Code int    `json:"code"`
	}
	Data string `json:"data"`
}

type acfunStreamsData struct {
	Stream []struct {
		StreamType string `json:"stream_type"`
		Resolution string `json:"resolution"`
		TotalSize  int64  `json:"total_size"`
		SliceNum   int    `json:"slice_num"`
		Segs       []struct {
			Url  string `json:"url"`
			Size int64  `json:"size"`
		}
		Width    string `json:"width"`
		Size     int    `json:"size"`
		Duration int    `json:"duration"`
	}
}
