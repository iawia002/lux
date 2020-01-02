package tiktok

import (
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/extractors"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

// Extract is the main function for extracting data
func Extract(uri string) ([]downloader.Data, error) {
	html, err := request.Get(uri, uri, nil)
	if err != nil {
		return nil, err
	}

	// There are a few json objects loaded into the html that are useful. We're able to parse the video url from the
	// videoObject json.

	videoScriptTag := utils.MatchOneOf(html, `<script type="application\/ld\+json" id="videoObject">(.*?)<\/script>`)
	if videoScriptTag == nil || len(videoScriptTag) < 2 {
		return nil, extractors.ErrURLParseFailed
	}
	videoJSON := videoScriptTag[1]
	videoURL := utils.GetStringFromJson(videoJSON, "contentUrl")

	// We can receive the title directly from this __NEXT_DATA__ object.

	nextScriptTag := utils.MatchOneOf(html, `<script id="__NEXT_DATA__" type="application\/json" crossorigin="anonymous">(.*?)<\/script>`)
	if nextScriptTag == nil || len(nextScriptTag) < 2 {
		return nil, extractors.ErrURLParseFailed
	}
	nextJSON := nextScriptTag[1]
	title := utils.GetStringFromJson(nextJSON, "props.pageProps.videoData.itemInfos.text")

	streams := map[string]downloader.Stream{}

	size, err := request.Size(videoURL, uri)
	if err != nil {
		return nil, err
	}
	urlData := downloader.URL{
		URL:  videoURL,
		Size: size,
		Ext:  "mp4",
	}
	streams["default"] = downloader.Stream{
		URLs: []downloader.URL{urlData},
		Size: size,
	}

	return []downloader.Data{
		{
			Site:    "TikTok tiktok.com",
			Title:   title,
			Type:    "video",
			Streams: streams,
			URL:     uri,
		},
	}, nil
}
