package ixigua

type ssrHydratedData struct {
	AnyVideo struct {
		GidInformation struct {
			PackerData struct {
				Video struct {
					Title         string `json:"title"`
					VideoResource struct {
						Normal struct {
							VideoList struct {
								Video1 video `json:"video_1"`
								Video2 video `json:"video_2"`
								Video3 video `json:"video_3"`
								Video4 video `json:"video_4"`
							} `json:"video_list"`
						} `json:"normal"`
						Dash120Fps struct {
							DynamicVideo struct {
								DynamicVideoList []video `json:"dynamic_video_list"`
								DynamicAudioList []video `json:"dynamic_audio_list"`
							} `json:"dynamic_video"`
						} `json:"dash_120fps"`
					} `json:"videoResource"`
				} `json:"video"`
			} `json:"packerData"`
		} `json:"gidInformation"`
	} `json:"anyVideo"`
}

type ssrHydratedDataEpisode struct {
	AnyVideo struct {
		GidInformation struct {
			PackerData struct {
				EpisodeInfo struct {
					Title string `json:"title"`
					Name  string `json:"name"`
				} `json:"episodeInfo"`
				VideoResource struct {
					Normal struct {
						VideoList struct {
							Video1 video `json:"video_1"`
							Video2 video `json:"video_2"`
							Video3 video `json:"video_3"`
							Video4 video `json:"video_4"`
						} `json:"video_list"`
					} `json:"normal"`
				} `json:"videoResource"`
			} `json:"packerData"`
		} `json:"gidInformation"`
	} `json:"anyVideo"`
}

type video struct {
	Definition string `json:"definition"`
	MainURL    string `json:"main_url"`
}
