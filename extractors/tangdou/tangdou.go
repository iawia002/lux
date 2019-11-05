package tangdou

import (
	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/extractors"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

const referer = "http://www.tangdou.com/html/playlist/view/4173"

// Extract is the main function for extracting data
func Extract(uri string) ([]downloader.Data, error) {
	if !config.Playlist {
		return []downloader.Data{tangdouDownload(uri)}, nil
	}
	html, err := request.Get(uri, referer, nil)
	if err != nil {
		return nil, err
	}
	videoIDs := utils.MatchAll(html, `<a target="tdplayer" href="(.+?)" class="title">`)
	needDownloadItems := utils.NeedDownloadList(len(videoIDs))
	extractedData := make([]downloader.Data, len(needDownloadItems))
	wgp := utils.NewWaitGroupPool(config.ThreadNumber)
	dataIndex := 0
	for index, videoID := range videoIDs {
		if !utils.ItemInSlice(index+1, needDownloadItems) || len(videoID) < 2 {
			continue
		}
		wgp.Add()
		go func(index int, videURI string, extractedData []downloader.Data) {
			defer wgp.Done()
			extractedData[index] = tangdouDownload(videURI)
		}(dataIndex, videoID[1], extractedData)
		dataIndex++
	}
	wgp.Wait()
	return extractedData, nil
}

// tangdouDownload download function for single url
func tangdouDownload(uri string) downloader.Data {
	html, err := request.Get(uri, referer, nil)
	if err != nil {
		return downloader.EmptyData(uri, err)
	}

	titles := utils.MatchOneOf(
		html, `<div class="title">(.+?)</div>`, `<meta name="description" content="(.+?)"`, `<title>(.+?)</title>`,
	)
	if titles == nil || len(titles) < 2 {
		return downloader.EmptyData(uri, extractors.ErrURLParseFailed)
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
			return downloader.EmptyData(uri, extractors.ErrURLParseFailed)
		}
		shareURL := shareURLs[1]

		signedVideo, err := request.Get(shareURL, uri, nil)
		if err != nil {
			return downloader.EmptyData(uri, err)
		}

		realURLs := utils.MatchOneOf(
			signedVideo, `src=\\"(.+?)\\"`,
		)
		if realURLs == nil || len(realURLs) < 2 {
			return downloader.EmptyData(uri, extractors.ErrURLParseFailed)
		}
		realURL = realURLs[1]
	} else {
		if len(videoURLs) < 2 {
			return downloader.EmptyData(uri, extractors.ErrURLParseFailed)
		}
		realURL = videoURLs[1]
	}

	size, err := request.Size(realURL, uri)
	if err != nil {
		return downloader.EmptyData(uri, err)
	}

	streams := map[string]downloader.Stream{
		"default": {
			URLs: []downloader.URL{
				{
					URL:  realURL,
					Size: size,
					Ext:  "mp4",
				},
			},
			Size: size,
		},
	}

	return downloader.Data{
		Site:    "糖豆广场舞 tangdou.com",
		Title:   title,
		Type:    "video",
		Streams: streams,
		URL:     uri,
	}
}
