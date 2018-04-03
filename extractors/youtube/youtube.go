package youtube

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type args struct {
	Title  string `json:"title"`
	Stream string `json:"url_encoded_fmt_stream_map"`
	Audio  string `json:"adaptive_fmts"`
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

func getSig(sig, js string) string {
	url := fmt.Sprintf("https://www.youtube.com%s", js)
	tokens, ok := tokensCache[url]
	if !ok {
		tokens = getSigTokens(request.Get(url, referer))
		tokensCache[url] = tokens
	}
	return decipherTokens(tokens, sig)
}

func genSignedURL(streamURL string, stream url.Values, js string) string {
	var realURL, sig string
	if strings.Contains(streamURL, "signature=") {
		// URL itself already has a signature parameter
		realURL = streamURL
	} else {
		// URL has no signature parameter
		sig = stream.Get("sig")
		if sig == "" {
			// Signature need decrypt
			sig = getSig(stream.Get("s"), js)
		}
		realURL = fmt.Sprintf("%s&signature=%s", streamURL, sig)
	}
	return realURL
}

// Download YouTube main download function
func Download(uri string) {
	if !config.Playlist {
		youtubeDownload(uri)
		return
	}
	listID := utils.MatchOneOf(uri, `(list|p)=([^/&]+)`)[2]
	if listID == "" {
		log.Fatal("Can't get list ID from URL")
	}
	html := request.Get("https://www.youtube.com/playlist?list="+listID, referer)
	// "videoId":"OQxX8zgyzuM","thumbnail"
	videoIDs := utils.MatchAll(html, `"videoId":"([^,]+?)","thumbnail"`)
	for _, videoID := range videoIDs {
		u := fmt.Sprintf(
			"https://www.youtube.com/watch?v=%s&list=%s", videoID[1], listID,
		)
		youtubeDownload(u)
	}
}

func youtubeDownload(uri string) downloader.VideoData {
	vid := utils.MatchOneOf(
		uri,
		`watch\?v=([^/&]+)`,
		`youtu\.be/([^?/]+)`,
		`embed/([^/?]+)`,
		`v/([^/?]+)`,
	)
	if vid == nil {
		log.Fatal("Can't find vid")
	}
	videoURL := fmt.Sprintf(
		"https://www.youtube.com/watch?v=%s&gl=US&hl=en&has_verified=1&bpctr=9999999999",
		vid[1],
	)
	html := request.Get(videoURL, referer)
	ytplayer := utils.MatchOneOf(html, `;ytplayer\.config\s*=\s*({.+?});`)[1]
	var youtube youtubeData
	json.Unmarshal([]byte(ytplayer), &youtube)
	title := youtube.Args.Title
	streams := strings.Split(youtube.Args.Stream, ",")

	format := extractVideoURLS(streams, uri, youtube.Assets.JS)

	// Audio only file
	for _, s := range strings.Split(youtube.Args.Audio, ",") {
		stream, _ := url.ParseQuery(s)
		if strings.HasPrefix(stream.Get("type"), "audio/mp4") {
			audioURL := genSignedURL(stream.Get("url"), stream, youtube.Assets.JS) + "&ratebypass=yes"
			size := request.Size(audioURL, uri)
			urlData := downloader.URLData{
				URL:  audioURL,
				Size: size,
				Ext:  "m4a",
			}
			format["audio"] = downloader.FormatData{
				URLs:    []downloader.URLData{urlData},
				Size:    size,
				Quality: stream.Get("type"),
			}
			break
		}
	}

	extractedData := downloader.VideoData{
		Site:    "YouTube youtube.com",
		Title:   utils.FileName(title),
		Type:    "video",
		Formats: format,
	}
	extractedData.Download(uri)
	return extractedData
}

func extractVideoURLS(streams []string, referer, assest string) map[string]downloader.FormatData {
	format := map[string]downloader.FormatData{}

	bestQualityURL, _ := url.ParseQuery(streams[0])
	bestQualityItag := bestQualityURL.Get("itag")

	for _, s := range streams {
		stream, _ := url.ParseQuery(s)
		itag := stream.Get("itag")

		if !utils.NeedExtract(itag, bestQualityItag) {
			continue
		}

		quality := stream.Get("quality")
		ext := utils.MatchOneOf(stream.Get("type"), `video/(\w+);`)[1]
		streamURL := stream.Get("url")
		realURL := genSignedURL(streamURL, stream, assest)
		size := request.Size(realURL, referer)
		urlData := downloader.URLData{
			URL:  realURL,
			Size: size,
			Ext:  ext,
		}
		format[itag] = downloader.FormatData{
			URLs:    []downloader.URLData{urlData},
			Size:    size,
			Quality: quality,
		}
	}

	format["default"] = format[bestQualityItag]
	delete(format, bestQualityItag)
	return format
}
