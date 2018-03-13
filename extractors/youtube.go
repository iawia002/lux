package extractors

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type args struct {
	Title  string `json:"title"`
	Stream string `json:"url_encoded_fmt_stream_map"`
}

type assets struct {
	JS string `json:"js"`
}

type youtubeData struct {
	Args   args   `json:"args"`
	Assets assets `json:"assets"`
}

func getSig(sig, js string) string {
	html := request.Get(fmt.Sprintf("https://www.youtube.com%s", js))
	return decipherTokens(getSigTokens(html), sig)
}

// Youtube download function
func Youtube(uri string) downloader.VideoData {
	patterns := []string{
		`watch\?v=(\w+)`,
		`youtu\.be/([^?/]+)`,
		`embed/([^/?]+)`,
		`v/([^/?]+)`,
	}
	vid := utils.MatchOneOf(patterns, uri)
	if vid == nil {
		log.Fatal("Can't find vid")
	}
	videoURL := fmt.Sprintf(
		"https://www.youtube.com/watch?v=%s&gl=US&hl=en&has_verified=1&bpctr=9999999999",
		vid[1],
	)
	html := request.Get(videoURL)
	ytplayer := utils.Match1(`;ytplayer\.config\s*=\s*({.+?});`, html)[1]
	var youtube youtubeData
	json.Unmarshal([]byte(ytplayer), &youtube)
	title := youtube.Args.Title
	streams := strings.Split(youtube.Args.Stream, ",")
	stream, _ := url.ParseQuery(streams[0]) // Best quality
	quality := stream.Get("quality")
	ext := utils.Match1(`video/(\w+);`, stream.Get("type"))[1]
	sig := stream.Get("sig")
	if sig == "" {
		sig = getSig(stream.Get("s"), youtube.Assets.JS)
	}
	realURL := fmt.Sprintf("%s&signature=%s", stream.Get("url"), sig)
	size := request.Size(realURL, uri)
	urlData := downloader.URLData{
		URL:  realURL,
		Size: size,
		Ext:  ext,
	}
	data := downloader.VideoData{
		Site:    "YouTube youtube.com",
		Title:   title,
		Type:    "video",
		URLs:    []downloader.URLData{urlData},
		Size:    size,
		Quality: quality,
	}
	data.Download(uri)
	return data
}
