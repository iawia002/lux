package youtube

type playerResponseData struct {
	PlayabilityStatus struct {
		Status          string `json:"status"`
		Reason          string `json:"reason"`
		PlayableInEmbed bool   `json:"playableInEmbed"`
		Miniplayer      struct {
			MiniplayerRenderer struct {
				PlaybackMode string `json:"playbackMode"`
			} `json:"miniplayerRenderer"`
		} `json:"miniplayer"`
		ContextParams string `json:"contextParams"`
	} `json:"playabilityStatus"`
	StreamingData struct {
		ExpiresInSeconds string   `json:"expiresInSeconds"`
		Formats          []Format `json:"formats"`
		AdaptiveFormats  []Format `json:"adaptiveFormats"`
		DashManifestURL  string   `json:"dashManifestUrl"`
		HlsManifestURL   string   `json:"hlsManifestUrl"`
	} `json:"streamingData"`
	VideoDetails struct {
		VideoID          string   `json:"videoId"`
		Title            string   `json:"title"`
		LengthSeconds    string   `json:"lengthSeconds"`
		Keywords         []string `json:"keywords"`
		ChannelID        string   `json:"channelId"`
		IsOwnerViewing   bool     `json:"isOwnerViewing"`
		ShortDescription string   `json:"shortDescription"`
		IsCrawlable      bool     `json:"isCrawlable"`
		Thumbnail        struct {
			Thumbnails []Thumbnail `json:"thumbnails"`
		} `json:"thumbnail"`
		AverageRating     float64 `json:"averageRating"`
		AllowRatings      bool    `json:"allowRatings"`
		ViewCount         string  `json:"viewCount"`
		Author            string  `json:"author"`
		IsPrivate         bool    `json:"isPrivate"`
		IsUnpluggedCorpus bool    `json:"isUnpluggedCorpus"`
		IsLiveContent     bool    `json:"isLiveContent"`
	} `json:"videoDetails"`
	Microformat struct {
		PlayerMicroformatRenderer struct {
			Thumbnail struct {
				Thumbnails []struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"thumbnails"`
			} `json:"thumbnail"`
			Title struct {
				SimpleText string `json:"simpleText"`
			} `json:"title"`
			Description struct {
				SimpleText string `json:"simpleText"`
			} `json:"description"`
			LengthSeconds      string   `json:"lengthSeconds"`
			OwnerProfileURL    string   `json:"ownerProfileUrl"`
			ExternalChannelID  string   `json:"externalChannelId"`
			IsFamilySafe       bool     `json:"isFamilySafe"`
			AvailableCountries []string `json:"availableCountries"`
			IsUnlisted         bool     `json:"isUnlisted"`
			HasYpcMetadata     bool     `json:"hasYpcMetadata"`
			ViewCount          string   `json:"viewCount"`
			Category           string   `json:"category"`
			PublishDate        string   `json:"publishDate"`
			OwnerChannelName   string   `json:"ownerChannelName"`
			UploadDate         string   `json:"uploadDate"`
		} `json:"playerMicroformatRenderer"`
	} `json:"microformat"`
}

type Format struct {
	ItagNo           int    `json:"itag"`
	URL              string `json:"url"`
	MimeType         string `json:"mimeType"`
	Quality          string `json:"quality"`
	Cipher           string `json:"signatureCipher"`
	Bitrate          int    `json:"bitrate"`
	FPS              int    `json:"fps"`
	Width            int    `json:"width"`
	Height           int    `json:"height"`
	LastModified     string `json:"lastModified"`
	ContentLength    int64  `json:"contentLength,string"`
	QualityLabel     string `json:"qualityLabel"`
	ProjectionType   string `json:"projectionType"`
	AverageBitrate   int    `json:"averageBitrate"`
	AudioQuality     string `json:"audioQuality"`
	ApproxDurationMs string `json:"approxDurationMs"`
	AudioSampleRate  string `json:"audioSampleRate"`
	AudioChannels    int    `json:"audioChannels"`

	// InitRange is only available for adaptive formats
	InitRange *struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"initRange"`

	// IndexRange is only available for adaptive formats
	IndexRange *struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"indexRange"`
}

type Thumbnails []Thumbnail

type Thumbnail struct {
	URL    string
	Width  uint
	Height uint
}
