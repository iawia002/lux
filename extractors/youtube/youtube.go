package youtube

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/rylio/ytdl"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/extractors"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type streamFormat struct {
	Itag          int    `json:"itag"`
	URL           string `json:"url"`
	MimeType      string `json:"mimeType"`
	ContentLength string `json:"contentLength"`
	QualityLabel  string `json:"qualityLabel"`
	AudioQuality  string `json:"audioQuality"`
}

type playerResponseType struct {
	StreamingData struct {
		Formats         []streamFormat  `json:"formats"`
		AdaptiveFormats adaptiveFormats `json:"adaptiveFormats"`
	} `json:"streamingData"`
	VideoDetails struct {
		Title string `json:"title"`
	} `json:"videoDetails"`
}

type adaptiveFormats []streamFormat

func (playerAdaptiveFormats adaptiveFormats) filterPlayerAdaptiveFormats(videoInfoFormats ytdl.FormatList) (filter adaptiveFormats) {
	videoInfoFormatMap := make(map[int]struct{}, len(videoInfoFormats))
	for _, f := range videoInfoFormats {
		videoInfoFormatMap[f.Number] = struct{}{}
	}
	for _, f := range playerAdaptiveFormats {
		if _, ok := videoInfoFormatMap[f.Itag]; ok {
			filter = append(filter, f)
		}
	}
	return
}

type youtubeData struct {
	Args struct {
		PlayerResponse string `json:"player_response"`
	} `json:"args"`
}

const referer = "https://www.youtube.com"

// Extract is the main function for extracting data
func Extract(uri string) ([]downloader.Data, error) {
	var err error
	if !config.Playlist {
		return []downloader.Data{youtubeDownload(uri)}, nil
	}
	listIDs := utils.MatchOneOf(uri, `(list|p)=([^/&]+)`)
	if listIDs == nil || len(listIDs) < 3 {
		return nil, extractors.ErrURLParseFailed
	}
	listID := listIDs[2]
	if len(listID) == 0 {
		return nil, errors.New("can't get list ID from URL")
	}

	html, err := request.Get("https://www.youtube.com/playlist?list="+listID, referer, nil)
	if err != nil {
		return nil, err
	}
	// "videoId":"OQxX8zgyzuM","thumbnail"
	videoIDs := utils.MatchAll(html, `"videoId":"([^,]+?)","thumbnail"`)
	needDownloadItems := utils.NeedDownloadList(len(videoIDs))
	extractedData := make([]downloader.Data, len(needDownloadItems))
	wgp := utils.NewWaitGroupPool(config.ThreadNumber)
	dataIndex := 0
	for index, videoID := range videoIDs {
		if !utils.ItemInSlice(index+1, needDownloadItems) || len(videoID) < 2 {
			continue
		}
		u := fmt.Sprintf(
			"https://www.youtube.com/watch?v=%s&list=%s", videoID[1], listID,
		)
		wgp.Add()
		go func(index int, u string, extractedData []downloader.Data) {
			defer wgp.Done()
			extractedData[index] = youtubeDownload(u)
		}(dataIndex, u, extractedData)
		dataIndex++
	}
	wgp.Wait()
	return extractedData, nil
}

// youtubeDownload download function for single url
func youtubeDownload(uri string) downloader.Data {
	vid := utils.MatchOneOf(
		uri,
		`watch\?v=([^/&]+)`,
		`youtu\.be/([^?/]+)`,
		`embed/([^/?]+)`,
		`v/([^/?]+)`,
	)
	if vid == nil || len(vid) < 2 {
		return downloader.EmptyData(uri, errors.New("can't find vid"))
	}

	videoURL := fmt.Sprintf(
		"https://www.youtube.com/watch?v=%s",
		vid[1],
	)

	videoInfo, err := ytdl.GetVideoInfo(uri)
	if err != nil {
		return downloader.EmptyData(uri, err)
	}

	html, err := request.Get(videoURL, referer, nil)
	if err != nil {
		return downloader.EmptyData(uri, err)
	}
	ytplayer := utils.MatchOneOf(html, `;ytplayer\.config\s*=\s*({.+?});`)
	if ytplayer == nil || len(ytplayer) < 2 {
		if strings.Contains(html, "LOGIN_REQUIRED") ||
			strings.Contains(html, "Sign in to confirm your age") {
			return downloader.EmptyData(uri, extractors.ErrLoginRequired)
		}
		return downloader.EmptyData(uri, extractors.ErrURLParseFailed)
	}

	var data youtubeData
	if err = json.Unmarshal([]byte(ytplayer[1]), &data); err != nil {
		return downloader.EmptyData(uri, err)
	}
	var playerResponse playerResponseType
	if err = json.Unmarshal([]byte(data.Args.PlayerResponse), &playerResponse); err != nil {
		return downloader.EmptyData(uri, err)
	}
	title := playerResponse.VideoDetails.Title
	playerResponse.StreamingData.AdaptiveFormats = playerResponse.
		StreamingData.
		AdaptiveFormats.
		filterPlayerAdaptiveFormats(videoInfo.Formats)

	streams, err := extractVideoURLS(playerResponse, videoInfo)
	if err != nil {
		return downloader.EmptyData(uri, err)
	}

	return downloader.Data{
		Site:    "YouTube youtube.com",
		Title:   title,
		Type:    "video",
		Streams: streams,
		URL:     uri,
	}
}

func getStreamExt(streamType string) string {
	// video/webm; codecs="vp8.0, vorbis" --> webm
	exts := utils.MatchOneOf(streamType, `(\w+)/(\w+);`)
	if exts == nil || len(exts) < 3 {
		return ""
	}
	return exts[2]
}

func getRealURL(videoFormat streamFormat, videoInfo *ytdl.VideoInfo, ext string) (*downloader.URL, error) {
	var ytdlFormat *ytdl.Format
	for _, f := range videoInfo.Formats {
		if f.Itag.Number == videoFormat.Itag {
			ytdlFormat = f
			break
		}
	}

	if ytdlFormat == nil {
		return nil, fmt.Errorf("unable to get info for itag %d", videoFormat.Itag)
	}

	realURL, err := videoInfo.GetDownloadURL(ytdlFormat)
	if err != nil {
		return nil, err
	}
	size, _ := strconv.ParseInt(videoFormat.ContentLength, 10, 64)
	return &downloader.URL{
		URL:  realURL.String(),
		Size: size,
		Ext:  ext,
	}, nil
}

func genStream(videoFormat streamFormat, videoInfo *ytdl.VideoInfo) (*downloader.Stream, error) {
	streamType := videoFormat.MimeType
	ext := getStreamExt(streamType)
	if ext == "" {
		return nil, fmt.Errorf("unable to get file extension of MimeType %s", streamType)
	}

	video, err := getRealURL(videoFormat, videoInfo, ext)
	if err != nil {
		return nil, err
	}

	var quality string
	if videoFormat.QualityLabel != "" {
		quality = fmt.Sprintf("%s %s", videoFormat.QualityLabel, streamType)
	} else {
		quality = streamType
	}

	return &downloader.Stream{
		URLs:    []downloader.URL{*video},
		Quality: quality,
	}, nil
}

func extractVideoURLS(data playerResponseType, videoInfo *ytdl.VideoInfo) (map[string]downloader.Stream, error) {
	streams := make(map[string]downloader.Stream, len(data.StreamingData.Formats)+len(data.StreamingData.AdaptiveFormats))
	for _, f := range data.StreamingData.Formats {
		stream, err := genStream(f, videoInfo)
		if err != nil {
			return nil, err
		}

		streams[strconv.Itoa(f.Itag)] = *stream
	}

	// Unlike `url_encoded_fmt_stream_map`, all videos in `adaptive_fmts` have no sound,
	// we need download video and audio both and then merge them.

	// Get separate m4a and webm audio streams for videos in AdaptiveFormats.
	// Prefer medium quality audio over low quality (there is no AUDIO_QUALITY_HIGH).
	var fM4aMedium, fM4aLow, fWebmMedium, fWebmLow *streamFormat
	for i, f := range data.StreamingData.AdaptiveFormats {
		switch {
		case strings.HasPrefix(f.MimeType, "audio/mp4"):
			if f.AudioQuality == "AUDIO_QUALITY_MEDIUM" {
				fM4aMedium = &data.StreamingData.AdaptiveFormats[i]
			} else {
				fM4aLow = &data.StreamingData.AdaptiveFormats[i]
			}
		case strings.HasPrefix(f.MimeType, "audio/webm"):
			if f.AudioQuality == "AUDIO_QUALITY_MEDIUM" {
				fWebmMedium = &data.StreamingData.AdaptiveFormats[i]
			} else {
				fWebmLow = &data.StreamingData.AdaptiveFormats[i]
			}
		}

		if fM4aMedium != nil && fWebmMedium != nil {
			break
		}
	}

	var audioM4a downloader.URL
	if fM4aMedium != nil {
		audioURL, err := getRealURL(*fM4aMedium, videoInfo, "m4a")
		if err != nil {
			return nil, err
		}
		audioM4a = *audioURL
	} else if fM4aLow != nil {
		audioURL, err := getRealURL(*fM4aLow, videoInfo, "m4a")
		if err != nil {
			return nil, err
		}
		audioM4a = *audioURL
	}

	var audioWebm downloader.URL
	if fWebmMedium != nil {
		audioURL, err := getRealURL(*fWebmMedium, videoInfo, "webm")
		if err != nil {
			return nil, err
		}
		audioWebm = *audioURL
	} else if fM4aLow != nil {
		audioURL, err := getRealURL(*fWebmLow, videoInfo, "webm")
		if err != nil {
			return nil, err
		}
		audioWebm = *audioURL
	}

	var emptyURL downloader.URL
	for _, f := range data.StreamingData.AdaptiveFormats {
		stream, err := genStream(f, videoInfo)
		if err != nil {
			return nil, err
		}

		// append audio stream only for adaptive video streams (not audio)
		switch {
		case strings.HasPrefix(f.MimeType, "video/mp4"):
			if audioM4a != emptyURL {
				stream.URLs = append(stream.URLs, audioM4a)
			}
		case strings.HasPrefix(f.MimeType, "video/webm"):
			if audioWebm != emptyURL {
				stream.URLs = append(stream.URLs, audioWebm)
			}
		}

		streams[strconv.Itoa(f.Itag)] = *stream
	}

	return streams, nil
}
