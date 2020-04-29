package qingting

import (
	"encoding/json"
	"fmt"
	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/extractors"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
	"io/ioutil"
	"net/http"
	"strings"
)

type ChannelInfoApi struct {
	Data ChannelInfo
	Code int
}

type ChannelInfo struct {
	ProgramCount int `json:"program_count"`
	Name         string
}

type ChannelAudioInfoApi struct {
	Data  []AudioInfo
	Code  int
	Total int
}

type AudioInfo struct {
	FilePath   string `json:"file_path"`
	Name       string
	ResId      int    `json:"res_id"`
	UpdateTime string `json:"update_time"`
	Duration   int
	Playcount  string
	Id         int
	Desc       string
	ChannelId  string `json:"channel_id"`
	Type       string
	ImgUrl     string `json:"img_url"`
}

// Extract is the main function for extracting data
func Extract(uri string) ([]downloader.Data, error) {
	channelId := extractChannelId(uri)

	// get info of the channel
	channelInfoUrl := getChannelInfoUrl(channelId)
	channelInfoResponse, err := http.Get(channelInfoUrl)
	if err != nil {
		fmt.Println("Error in fetching JSON")
		return nil, extractors.ErrURLParseFailed
	}
	defer channelInfoResponse.Body.Close()
	channelInfoBody, err := ioutil.ReadAll(channelInfoResponse.Body)
	var parsedChannelJson ChannelInfoApi
	json.Unmarshal(channelInfoBody, &parsedChannelJson)

	// request API and parse it
	audioInfoUrl := getChannelAudioInfoUrl(channelId)
	response, err := http.Get(audioInfoUrl)
	if err != nil {
		fmt.Println("Error in fetching JSON")
		return nil, err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	var parsedJson ChannelAudioInfoApi
	json.Unmarshal(body, &parsedJson)

	// handle playlist
	needDownloadItems := utils.NeedDownloadList(len(parsedJson.Data))
	extractedData := make([]downloader.Data, len(parsedJson.Data))
	wgp := utils.NewWaitGroupPool(config.ThreadNumber)
	dataIndex := 0
	for index, audioInfo := range parsedJson.Data {

		if !utils.ItemInSlice(index+1, needDownloadItems) {
			continue
		}
		wgp.Add()

		go func(index int, audioInfo AudioInfo, extractedData []downloader.Data) {
			defer wgp.Done()
			extractedData[index] = qingtingDownload(audioInfo, uri)
		}(dataIndex, audioInfo, extractedData)
		dataIndex++

	}
	wgp.Wait()
	return extractedData, nil
}

func extractChannelId(uri string) string {
	s := strings.Split(uri, "/")
	channid := s[len(s)-1]
	return channid
}

func getChannelInfoUrl(channelId string) string {
	return "http://i.qingting.fm/wapi/channels/" + channelId
}

func getChannelAudioInfoUrl(channelId string) string {
	return "http://i.qingting.fm/wapi/channels/" + channelId + "/programs/page/1/pagesize/250"
}

func getAudioFilePath(filePath string) string {
	return "http://od.qingting.fm/" + filePath
}

func qingtingDownload(audioInfo AudioInfo, uri string) downloader.Data {
	title := audioInfo.Name
	audioPath := getAudioFilePath(audioInfo.FilePath)

	streams := map[string]downloader.Stream{}

	size, err := request.Size(audioPath, uri)
	if err != nil {
		return downloader.EmptyData(uri, extractors.ErrURLParseFailed)
	}
	urlData := downloader.URL{
		URL:  audioPath,
		Size: size,
		Ext:  "m4a",
	}
	streams["default"] = downloader.Stream{
		URLs: []downloader.URL{urlData},
		Size: size,
	}
	return downloader.Data{
		Site:    "qingting fm",
		Title:   title,
		Type:    "aduio",
		Streams: streams,
		URL:     uri,
	}

}
