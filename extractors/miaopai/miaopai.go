package miaopai

import (
	"encoding/json"
	"fmt"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type miaopai struct {
	Data struct {
		Description string `json:"description"`
		MetaData    []struct {
			URLs struct {
				M string `json:"m"`
			} `json:"play_urls"`
		} `json:"meta_data"`
	} `json:"data"`
}

// Download main download function
func Download(url string) ([]downloader.Data, error) {
	id := utils.MatchOneOf(url, `/media/([^\./]+)`, `/show(?:/channel)?/([^\./]+)`)[1]
	jsonString, err := request.Get(
		fmt.Sprintf("https://n.miaopai.com/api/aj_media/info.json?smid=%s", id), url, nil,
	)
	if err != nil {
		return downloader.EmptyList, err
	}
	var data miaopai
	json.Unmarshal([]byte(jsonString), &data)

	realURL := data.Data.MetaData[0].URLs.M
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
			Site:    "秒拍 miaopai.com",
			Title:   data.Data.Description,
			Type:    "video",
			Streams: streams,
			URL:     url,
		},
	}, nil
}
