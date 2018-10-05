package youtube

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type args struct {
	Title  string `json:"title"`
	Stream string `json:"adaptive_fmts"`
	// not every page has `adaptive_fmts` field https://youtu.be/DNaOZovrSVo
	Stream2 string `json:"url_encoded_fmt_stream_map"`
}

type assets struct {
	JS string `json:"js"`
}

type youtubeData struct {
	Args   args   `json:"args"`
	Assets assets `json:"assets"`
}

const referer = "https://www.youtube.com"

var tokensCache = make(map[string][]string)

func getSig(sig, js string) (string, error) {
	sigURL := fmt.Sprintf("https://www.youtube.com%s", js)
	tokens, ok := tokensCache[sigURL]
	if !ok {
		html, err := request.Get(sigURL, referer, nil)
		if err != nil {
			return "", err
		}
		tokens, err = getSigTokens(html)
		if err != nil {
			return "", err
		}
		tokensCache[sigURL] = tokens
	}
	return decipherTokens(tokens, sig), nil
}

func genSignedURL(streamURL string, stream url.Values, js string) (string, error) {
	var (
		realURL, sig string
		err          error
	)
	if strings.Contains(streamURL, "signature=") {
		// URL itself already has a signature parameter
		realURL = streamURL
	} else {
		// URL has no signature parameter
		sig = stream.Get("sig")
		if sig == "" {
			// Signature need decrypt
			sig, err = getSig(stream.Get("s"), js)
			if err != nil {
				return "", err
			}
		}
		realURL = fmt.Sprintf("%s&signature=%s", streamURL, sig)
	}
	if !strings.Contains(realURL, "ratebypass") {
		realURL += "&ratebypass=yes"
	}
	return realURL, nil
}

// Download YouTube main download function
func Download(uri string) ([]downloader.VideoData, error) {
	var err error
	if !config.Playlist {
		data, err := youtubeDownload(uri)
		if err != nil {
			return downloader.EmptyData, err
		}
		return []downloader.VideoData{data}, nil
	}
	listID := utils.MatchOneOf(uri, `(list|p)=([^/&]+)`)[2]
	if listID == "" {
		return downloader.EmptyData, errors.New("can't get list ID from URL")
	}
	html, err := request.Get("https://www.youtube.com/playlist?list="+listID, referer, nil)
	if err != nil {
		return downloader.EmptyData, err
	}
	// "videoId":"OQxX8zgyzuM","thumbnail"
	videoIDs := utils.MatchAll(html, `"videoId":"([^,]+?)","thumbnail"`)
	needDownloadItems := utils.NeedDownloadList(len(videoIDs))
	extractedData := make([]downloader.VideoData, len(needDownloadItems))
	wgp := utils.NewWaitGroupPool(config.ThreadNumber)
	for index, videoID := range videoIDs {
		if !utils.ItemInSlice(index+1, needDownloadItems) {
			continue
		}
		u := fmt.Sprintf(
			"https://www.youtube.com/watch?v=%s&list=%s", videoID[1], listID,
		)
		wgp.Add()
		go func(index int, u string, extractedData []downloader.VideoData) {
			defer wgp.Done()
			data, err := youtubeDownload(u)
			if err == nil {
				// if err is not nil, the data is empty struct
				extractedData[index] = data
			}
		}(index, u, extractedData)
	}
	wgp.Wait()
	return extractedData, nil
}

func youtubeDownload(uri string) (downloader.VideoData, error) {
	var err error
	vid := utils.MatchOneOf(
		uri,
		`watch\?v=([^/&]+)`,
		`youtu\.be/([^?/]+)`,
		`embed/([^/?]+)`,
		`v/([^/?]+)`,
	)
	if vid == nil {
		return downloader.VideoData{}, errors.New("can't find vid")
	}
	videoURL := fmt.Sprintf(
		"https://www.youtube.com/watch?v=%s&gl=US&hl=en&has_verified=1&bpctr=9999999999",
		vid[1],
	)
	html, err := request.Get(videoURL, referer, nil)
	if err != nil {
		return downloader.VideoData{}, err
	}
	ytplayer := utils.MatchOneOf(html, `;ytplayer\.config\s*=\s*({.+?});`)[1]
	var youtube youtubeData
	json.Unmarshal([]byte(ytplayer), &youtube)
	title := youtube.Args.Title

	format, err := extractVideoURLS(youtube, uri)
	if err != nil {
		return downloader.VideoData{}, err
	}

	return downloader.VideoData{
		Site:    "YouTube youtube.com",
		Title:   utils.FileName(title),
		Type:    "video",
		Formats: format,
	}, nil
}

func extractVideoURLS(data youtubeData, referer string) (map[string]downloader.FormatData, error) {
	streams := strings.Split(data.Args.Stream, ",")
	if data.Args.Stream == "" {
		streams = strings.Split(data.Args.Stream2, ",")
	}
	var ext string
	var audio downloader.URLData
	format := map[string]downloader.FormatData{}

	for _, s := range streams {
		stream, err := url.ParseQuery(s)
		if err != nil {
			return nil, err
		}
		itag := stream.Get("itag")
		streamType := stream.Get("type")
		isAudio := strings.HasPrefix(streamType, "audio/mp4")

		quality := stream.Get("quality_label")
		if quality == "" {
			quality = stream.Get("quality") // for url_encoded_fmt_stream_map
		}
		if quality != "" {
			quality = fmt.Sprintf("%s %s", quality, streamType)
		} else {
			quality = streamType
		}
		if isAudio {
			// audio file use m4a extension
			ext = "m4a"
		} else {
			ext = utils.MatchOneOf(streamType, `(\w+)/(\w+);`)[2]
		}
		realURL, err := genSignedURL(stream.Get("url"), stream, data.Assets.JS)
		if err != nil {
			return nil, err
		}
		size, err := request.Size(realURL, referer)
		if err != nil {
			return nil, err
		}
		urlData := downloader.URLData{
			URL:  realURL,
			Size: size,
			Ext:  ext,
		}
		if isAudio {
			// Audio data for merging with video
			audio = urlData
		}
		format[itag] = downloader.FormatData{
			URLs:    []downloader.URLData{urlData},
			Size:    size,
			Quality: quality,
		}
	}

	// `url_encoded_fmt_stream_map`
	if data.Args.Stream == "" {
		return format, nil
	}

	// Unlike `url_encoded_fmt_stream_map`, all videos in `adaptive_fmts` have no sound,
	// we need download video and audio both and then merge them.
	// Another problem is that even if we add `ratebypass=yes`, the download speed still slow sometimes. https://github.com/iawia002/annie/issues/191#issuecomment-405449649

	// All videos here have no sound and need to be added separately
	for itag, f := range format {
		if strings.Contains(f.Quality, "video/") {
			f.Size += audio.Size
			f.URLs = append(f.URLs, audio)
			format[itag] = f
		}
	}
	return format, nil
}
