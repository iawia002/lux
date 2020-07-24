package youtube

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/rylio/ytdl"

	"github.com/iawia002/annie/extractors/types"
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

type formats []*streamFormat

func (playerAdaptiveFormats formats) filterPlayerAdaptiveFormats(videoInfoFormats ytdl.FormatList) (filter formats) {
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

type playerResponseType struct {
	StreamingData struct {
		Formats         formats `json:"formats"`
		AdaptiveFormats formats `json:"adaptiveFormats"`
	} `json:"streamingData"`
	VideoDetails struct {
		Title string `json:"title"`
	} `json:"videoDetails"`
}

type youtubeData struct {
	Args struct {
		PlayerResponse string `json:"player_response"`
	} `json:"args"`
}

const referer = "https://www.youtube.com"

type extractor struct{}

// New returns a youtube extractor.
func New() types.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(uri string, option types.Options) ([]*types.Data, error) {
	var err error
	if !option.Playlist {
		return []*types.Data{youtubeDownload(uri)}, nil
	}
	listIDs := utils.MatchOneOf(uri, `(list|p)=([^/&]+)`)
	if listIDs == nil || len(listIDs) < 3 {
		return nil, types.ErrURLParseFailed
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
	needDownloadItems := utils.NeedDownloadList(option.Items, option.ItemStart, option.ItemEnd, len(videoIDs))
	extractedData := make([]*types.Data, len(needDownloadItems))
	wgp := utils.NewWaitGroupPool(option.ThreadNumber)
	dataIndex := 0
	for index, videoID := range videoIDs {
		if !utils.ItemInSlice(index+1, needDownloadItems) || len(videoID) < 2 {
			continue
		}
		u := fmt.Sprintf(
			"https://www.youtube.com/watch?v=%s&list=%s", videoID[1], listID,
		)
		wgp.Add()
		go func(index int, u string, extractedData []*types.Data) {
			defer wgp.Done()
			extractedData[index] = youtubeDownload(u)
		}(dataIndex, u, extractedData)
		dataIndex++
	}
	wgp.Wait()
	return extractedData, nil
}

// youtubeDownload download function for single url
func youtubeDownload(uri string) *types.Data {
	vid := utils.MatchOneOf(
		uri,
		`watch\?v=([^/&]+)`,
		`youtu\.be/([^?/]+)`,
		`embed/([^/?]+)`,
		`v/([^/?]+)`,
	)
	if vid == nil || len(vid) < 2 {
		return types.EmptyData(uri, errors.New("can't find vid"))
	}

	videoURL := fmt.Sprintf(
		"https://www.youtube.com/watch?v=%s",
		vid[1],
	)

	videoInfo, err := ytdl.DefaultClient.GetVideoInfo(context.TODO(), uri)
	if err != nil {
		return types.EmptyData(uri, err)
	}

	html, err := request.Get(videoURL, referer, nil)
	if err != nil {
		return types.EmptyData(uri, err)
	}
	ytplayer := utils.MatchOneOf(html, `;ytplayer\.config\s*=\s*({.+?});`)
	if ytplayer == nil || len(ytplayer) < 2 {
		if strings.Contains(html, "LOGIN_REQUIRED") ||
			strings.Contains(html, "Sign in to confirm your age") {
			return types.EmptyData(uri, types.ErrLoginRequired)
		}
		return types.EmptyData(uri, types.ErrURLParseFailed)
	}

	var data youtubeData
	if err = json.Unmarshal([]byte(ytplayer[1]), &data); err != nil {
		return types.EmptyData(uri, err)
	}
	var playerResponse playerResponseType
	if err = json.Unmarshal([]byte(data.Args.PlayerResponse), &playerResponse); err != nil {
		return types.EmptyData(uri, err)
	}
	title := playerResponse.VideoDetails.Title
	playerResponse.StreamingData.AdaptiveFormats = playerResponse.
		StreamingData.
		AdaptiveFormats.
		filterPlayerAdaptiveFormats(videoInfo.Formats)

	streams, err := extractVideoURLS(playerResponse, videoInfo)
	if err != nil {
		return types.EmptyData(uri, err)
	}

	return &types.Data{
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

func getRealURL(videoFormat *streamFormat, videoInfo *ytdl.VideoInfo, ext string) (*types.Part, error) {
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

	realURL, err := ytdl.DefaultClient.GetDownloadURL(context.TODO(), videoInfo, ytdlFormat)
	if err != nil {
		return nil, err
	}
	size, _ := strconv.ParseInt(videoFormat.ContentLength, 10, 64)
	return &types.Part{
		URL:  realURL.String(),
		Size: size,
		Ext:  ext,
	}, nil
}

func genStream(videoFormat *streamFormat, videoInfo *ytdl.VideoInfo) (*types.Stream, error) {
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

	return &types.Stream{
		ID:      strconv.Itoa(videoFormat.Itag),
		Parts:   []*types.Part{video},
		Quality: quality,
		NeedMux: true,
	}, nil
}

func extractVideoURLS(data playerResponseType, videoInfo *ytdl.VideoInfo) (map[string]*types.Stream, error) {
	streams := make(map[string]*types.Stream, len(data.StreamingData.Formats)+len(data.StreamingData.AdaptiveFormats))
	for _, f := range data.StreamingData.Formats {
		stream, err := genStream(f, videoInfo)
		if err != nil {
			return nil, err
		}

		streams[strconv.Itoa(f.Itag)] = stream
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
				fM4aMedium = data.StreamingData.AdaptiveFormats[i]
			} else {
				fM4aLow = data.StreamingData.AdaptiveFormats[i]
			}
		case strings.HasPrefix(f.MimeType, "audio/webm"):
			if f.AudioQuality == "AUDIO_QUALITY_MEDIUM" {
				fWebmMedium = data.StreamingData.AdaptiveFormats[i]
			} else {
				fWebmLow = data.StreamingData.AdaptiveFormats[i]
			}
		}

		if fM4aMedium != nil && fWebmMedium != nil {
			break
		}
	}

	var audioM4a *types.Part
	if fM4aMedium != nil {
		audioURL, err := getRealURL(fM4aMedium, videoInfo, "m4a")
		if err != nil {
			return nil, err
		}
		audioM4a = audioURL
	} else if fM4aLow != nil {
		audioURL, err := getRealURL(fM4aLow, videoInfo, "m4a")
		if err != nil {
			return nil, err
		}
		audioM4a = audioURL
	}

	var audioWebm *types.Part
	if fWebmMedium != nil {
		audioURL, err := getRealURL(fWebmMedium, videoInfo, "webm")
		if err != nil {
			return nil, err
		}
		audioWebm = audioURL
	} else if fWebmLow != nil {
		audioURL, err := getRealURL(fWebmLow, videoInfo, "webm")
		if err != nil {
			return nil, err
		}
		audioWebm = audioURL
	}

	for _, f := range data.StreamingData.AdaptiveFormats {
		stream, err := genStream(f, videoInfo)
		if err != nil {
			return nil, err
		}

		// append audio stream only for adaptive video streams (not audio)
		switch {
		case strings.HasPrefix(f.MimeType, "video/mp4"):
			if audioM4a != nil {
				stream.Parts = append(stream.Parts, audioM4a)
			}
		case strings.HasPrefix(f.MimeType, "video/webm"):
			if audioWebm != nil {
				stream.Parts = append(stream.Parts, audioWebm)
			}
		}

		streams[strconv.Itoa(f.Itag)] = stream
	}
	return streams, nil
}
