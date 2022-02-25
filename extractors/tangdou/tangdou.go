package tangdou

import (
	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	extractors.Register("tangdou", New())
}

const referer = "http://www.tangdou.com/html/playlist/view/4173"

type extractor struct{}

// New returns a tangdou extractor.
func New() extractors.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	if !option.Playlist {
		return []*extractors.Data{tangdouDownload(url)}, nil
	}

	html, err := request.Get(url, referer, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	videoIDs := utils.MatchAll(html, `<a target="tdplayer" href="(.+?)" class="title">`)
	needDownloadItems := utils.NeedDownloadList(option.Items, option.ItemStart, option.ItemEnd, len(videoIDs))
	extractedData := make([]*extractors.Data, len(needDownloadItems))
	wgp := utils.NewWaitGroupPool(option.ThreadNumber)
	dataIndex := 0
	for index, videoID := range videoIDs {
		if !utils.ItemInSlice(index+1, needDownloadItems) || len(videoID) < 2 {
			continue
		}
		wgp.Add()
		go func(index int, videURI string, extractedData []*extractors.Data) {
			defer wgp.Done()
			extractedData[index] = tangdouDownload(videURI)
		}(dataIndex, videoID[1], extractedData)
		dataIndex++
	}
	wgp.Wait()
	return extractedData, nil
}

// tangdouDownload download function for single url
func tangdouDownload(uri string) *extractors.Data {
	html, err := request.Get(uri, referer, nil)
	if err != nil {
		return extractors.EmptyData(uri, err)
	}

	titles := utils.MatchOneOf(
		html, `<div class="title">(.+?)</div>`, `<meta name="description" content="(.+?)"`, `<title>(.+?)</title>`,
	)
	if titles == nil || len(titles) < 2 {
		return extractors.EmptyData(uri, errors.WithStack(extractors.ErrURLParseFailed))
	}
	title := titles[1]

	var realURL string
	videoURLs := utils.MatchOneOf(
		html, `video:'(.+?)'`, `video:"(.+?)"`, `<video.*src="(.+?)"`,
	)
	if videoURLs == nil {
		shareURLs := utils.MatchOneOf(
			html, `<div class="video">\s*<script src="(.+?)"`,
		)
		if shareURLs == nil || len(shareURLs) < 2 {
			return extractors.EmptyData(uri, errors.WithStack(extractors.ErrURLParseFailed))
		}
		shareURL := shareURLs[1]

		signedVideo, err := request.Get(shareURL, uri, nil)
		if err != nil {
			return extractors.EmptyData(uri, err)
		}

		realURLs := utils.MatchOneOf(
			signedVideo, `src=\\"(.+?)\\"`,
		)
		if realURLs == nil || len(realURLs) < 2 {
			return extractors.EmptyData(uri, errors.WithStack(extractors.ErrURLParseFailed))
		}
		realURL = realURLs[1]
	} else {
		if len(videoURLs) < 2 {
			return extractors.EmptyData(uri, errors.WithStack(extractors.ErrURLParseFailed))
		}
		realURL = videoURLs[1]
	}

	size, err := request.Size(realURL, uri)
	if err != nil {
		return extractors.EmptyData(uri, err)
	}

	streams := map[string]*extractors.Stream{
		"default": {
			Parts: []*extractors.Part{
				{
					URL:  realURL,
					Size: size,
					Ext:  "mp4",
				},
			},
			Size: size,
		},
	}

	return &extractors.Data{
		Site:    "糖豆广场舞 tangdou.com",
		Title:   title,
		Type:    extractors.DataTypeVideo,
		Streams: streams,
		URL:     uri,
	}
}
