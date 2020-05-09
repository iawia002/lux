package tiktok

import (
	"github.com/iawia002/annie/extractors/types"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type extractor struct{}

// New returns a youtube extractor.
func New() types.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option types.Options) ([]*types.Data, error) {
	html, err := request.Get(url, url, nil)
	if err != nil {
		return nil, err
	}

	// There are a few json objects loaded into the html that are useful. We're able to parse the video url from the
	// videoObject json.

	videoScriptTag := utils.MatchOneOf(html, `<script type="application\/ld\+json" id="videoObject">(.*?)<\/script>`)
	if videoScriptTag == nil || len(videoScriptTag) < 2 {
		return nil, types.ErrURLParseFailed
	}
	videoJSON := videoScriptTag[1]
	videoURL := utils.GetStringFromJSON(videoJSON, "contentUrl")

	// We can receive the title directly from this __NEXT_DATA__ object.

	nextScriptTag := utils.MatchOneOf(html, `<script id="__NEXT_DATA__" type="application\/json" crossorigin="anonymous">(.*?)<\/script>`)
	if nextScriptTag == nil || len(nextScriptTag) < 2 {
		return nil, types.ErrURLParseFailed
	}
	nextJSON := nextScriptTag[1]
	title := utils.GetStringFromJSON(nextJSON, "props.pageProps.videoData.itemInfos.text")

	streams := make(map[string]*types.Stream)

	size, err := request.Size(videoURL, url)
	if err != nil {
		return nil, err
	}
	urlData := &types.Part{
		URL:  videoURL,
		Size: size,
		Ext:  "mp4",
	}
	streams["default"] = &types.Stream{
		Parts: []*types.Part{urlData},
		Size:  size,
	}

	return []*types.Data{
		{
			Site:    "TikTok tiktok.com",
			Title:   title,
			Type:    types.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
