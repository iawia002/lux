package miaopai

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/extractors"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type miaopaiData struct {
	Data struct {
		Description string `json:"description"`
		MetaData    []struct {
			URLs struct {
				M string `json:"m"`
			} `json:"play_urls"`
		} `json:"meta_data"`
	} `json:"data"`
}

func getRandomString(l int) string {
	rand.Seed(time.Now().UnixNano())

	s := make([]string, 0)
	chars := []string{
		"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "n", "m", "o", "p", "q", "r", "s", "t", "u", "v",
		"w", "x", "y", "z", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
	}
	for i := 0; i < l; i++ {
		s = append(s, chars[rand.Intn(len(chars)-1)])
	}
	return strings.Join(s, "")
}

// Extract is the main function for extracting data
func Extract(url string) ([]downloader.Data, error) {
	ids := utils.MatchOneOf(url, `/media/([^\./]+)`, `/show(?:/channel)?/([^\./]+)`)
	if ids == nil || len(ids) < 2 {
		return nil, extractors.ErrURLParseFailed
	}
	id := ids[1]

	randomString := getRandomString(10)

	var data miaopaiData
	jsonString, err := request.Get(
		fmt.Sprintf("https://n.miaopai.com/api/aj_media/info.json?smid=%s&appid=530&_cb=_jsonp%s", id, randomString),
		url, nil,
	)
	if err != nil {
		return nil, err
	}

	match := utils.MatchOneOf(jsonString, randomString+`\((.*)\);$`)
	if match == nil || len(match) < 2 {
		return nil, errors.New("获取视频信息失败。")
	}

	err = json.Unmarshal([]byte(match[1]), &data)
	if err != nil {
		return nil, err
	}

	realURL := data.Data.MetaData[0].URLs.M
	size, err := request.Size(realURL, url)
	if err != nil {
		return nil, err
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
