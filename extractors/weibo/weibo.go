package weibo

import (
	"fmt"
	netURL "net/url"
	"strings"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/parser"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

func downloadWeiboTV(url string) ([]downloader.Data, error) {
	headers := map[string]string{
		"Cookie": "SUB=_2AkMsZ8xOf8NxqwJRmP4RzGLqbo5xyQDEieKaOz2VJRMxHRl-yj83qlEotRB6B-fiobWQ5vdEoYw7bCoCdf4KyP8O3Ujq",
	}
	html, err := request.Get(url, url, headers)
	if err != nil {
		return downloader.EmptyList, err
	}
	doc, err := parser.GetDoc(html)
	if err != nil {
		return downloader.EmptyList, err
	}
	title := strings.TrimSpace(
		strings.Replace(doc.Find(".info_txt").First().Text(), "\u200B", " ", -1), // Zero width space.
	)
	// http://f.us.sinaimg.cn/003Cddn4lx07oCX1hC0001040200hkQk0k010.mp4?label=mp4_hd&template=852x480.20&Expires=1541041515&ssig=%2BYnCmZaToS&KID=unistore,video
	// &480=http://f.us.sinaimg.cn/003Cddn4lx07oCX1hC0001040200hkQk0k010.mp4?label=mp4_hd&template=852x480.20&Expires=1541041515&ssig=%2BYnCmZaToS&KID=unistore,video
	// &720=http://f.us.sinaimg.cn/004cqzndlx07oCX1kMOQ01040200vyxj0k010.mp4?label=mp4_720p&template=1280x720.20&Expires=1541041515&ssig=Fdasnr1aW6&KID=unistore,video&qType=720
	realURL, err := netURL.PathUnescape(
		utils.MatchOneOf(html, `video-sources="fluency=(.+?)"`)[1],
	)
	if err != nil {
		return downloader.EmptyList, err
	}
	quality := []string{"480", "720"}
	streams := make(map[string]downloader.Stream, len(quality))
	for _, q := range quality {
		urlList := strings.Split(realURL, fmt.Sprintf("&%s=", q))
		u := urlList[len(urlList)-1]
		if !strings.HasPrefix(u, "http") {
			continue
		}
		size, err := request.Size(u, url)
		if err != nil {
			return downloader.EmptyList, err
		}
		streams[q] = downloader.Stream{
			URLs: []downloader.URL{
				{
					URL:  u,
					Size: size,
					Ext:  "mp4",
				},
			},
			Size:    size,
			Quality: q,
		}
	}
	return []downloader.Data{
		{
			Site:    "微博 weibo.com",
			Title:   title,
			Type:    "video",
			Streams: streams,
			URL:     url,
		},
	}, nil
}

// Extract is the main function for extracting data
func Extract(url string) ([]downloader.Data, error) {
	if !strings.Contains(url, "m.weibo.cn") {
		if strings.Contains(url, "weibo.com/tv/v/") {
			return downloadWeiboTV(url)
		}
		url = strings.Replace(url, "weibo.com", "m.weibo.cn", 1)
	}
	html, err := request.Get(url, url, nil)
	if err != nil {
		return downloader.EmptyList, err
	}
	title := utils.MatchOneOf(
		html, `"content2": "(.+?)",`, `"status_title": "(.+?)",`,
	)[1]
	realURL := utils.MatchOneOf(
		html, `"stream_url_hd": "(.+?)"`, `"stream_url": "(.+?)"`,
	)[1]
	size, err := request.Size(realURL, url)
	if err != nil {
		return downloader.EmptyList, err
	}
	urlData := downloader.URL{
		URL:  realURL,
		Size: size,
		Ext:  "mp4",
	}
	streams := map[string]downloader.Stream{
		"default": {
			URLs: []downloader.URL{urlData},
			Size: size,
		},
	}

	return []downloader.Data{
		{
			Site:    "微博 weibo.com",
			Title:   title,
			Type:    "video",
			Streams: streams,
			URL:     url,
		},
	}, nil
}
