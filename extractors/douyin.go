package extractors

import (
	"encoding/json"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/extractors/utils"
)

type DouyinVideoUrlData struct {
	Url_list []string
}

type DouyinVideoData struct {
	Play_addr      DouyinVideoUrlData
	Real_play_addr string
}

type DouyinData struct {
	Video DouyinVideoData
	Desc  string
}

func Douyin(url string) utils.VideoData {
	html := downloader.Get(url)
	vData := downloader.Match1(`var data = \[(.*?)\];`, html)[1]
	var dataDict DouyinData
	json.Unmarshal([]byte(vData), &dataDict)

	data := utils.VideoData{
		Site:  "抖音 douyin.com",
		Title: dataDict.Desc,
		Url:   dataDict.Video.Real_play_addr,
		Ext:   "mp4",
	}
	data.Size = downloader.UrlSize(data.Url)
	downloader.UlrSave(data)
	return data
}
