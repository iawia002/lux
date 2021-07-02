package youtube

import (
	"fmt"
	"strconv"

	"github.com/kkdai/youtube/v2"

	"github.com/iawia002/annie/extractors/types"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

const referer = "https://www.youtube.com"

type extractor struct {
	client *youtube.Client
}

// New returns a youtube extractor.
func New() types.Extractor {
	return &extractor{
		client: &youtube.Client{},
	}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option types.Options) ([]*types.Data, error) {
	if !option.Playlist {
		video, err := e.client.GetVideo(url)
		if err != nil {
			return nil, err
		}
		return []*types.Data{e.youtubeDownload(url, video)}, nil
	}

	playlist, err := e.client.GetPlaylist(url)
	if err != nil {
		return nil, err
	}

	needDownloadItems := utils.NeedDownloadList(option.Items, option.ItemStart, option.ItemEnd, len(playlist.Videos))
	extractedData := make([]*types.Data, len(needDownloadItems))
	wgp := utils.NewWaitGroupPool(option.ThreadNumber)
	dataIndex := 0
	for index, videoEntry := range playlist.Videos {
		if !utils.ItemInSlice(index+1, needDownloadItems) {
			continue
		}

		wgp.Add()
		go func(index int, extractedData []*types.Data) {
			defer wgp.Done()
			video, err := e.client.VideoFromPlaylistEntry(videoEntry)
			if err != nil {
				return
			}
			extractedData[index] = e.youtubeDownload(url, video)
		}(dataIndex, extractedData)
		dataIndex++
	}
	wgp.Wait()
	return extractedData, nil
}

// youtubeDownload download function for single url
func (e *extractor) youtubeDownload(url string, video *youtube.Video) *types.Data {
	streams := make(map[string]*types.Stream, len(video.Formats))
	audioCache := make(map[string]*types.Part)

	for i := range video.Formats {
		f := &video.Formats[i]
		itag := strconv.Itoa(f.ItagNo)
		quality := f.MimeType
		if f.QualityLabel != "" {
			quality = fmt.Sprintf("%s %s", f.QualityLabel, f.MimeType)
		}

		part, err := e.genPartByFormat(video, f)
		if err != nil {
			return types.EmptyData(url, err)
		}
		stream := &types.Stream{
			ID:      itag,
			Parts:   []*types.Part{part},
			Quality: quality,
			Ext:     part.Ext,
			NeedMux: true,
		}

		// Unlike `url_encoded_fmt_stream_map`, all videos in `adaptive_fmts` have no sound,
		// we need download video and audio both and then merge them.
		// video format with audio:
		//   AudioSampleRate: "44100", AudioChannels: 2
		// video format without audio:
		//   AudioSampleRate: "", AudioChannels: 0
		if f.AudioChannels == 0 {
			audioPart, ok := audioCache[part.Ext]
			if !ok {
				audio, err := getVideoAudio(video, part.Ext)
				if err != nil {
					return types.EmptyData(url, err)
				}
				audioPart, err = e.genPartByFormat(video, audio)
				if err != nil {
					return types.EmptyData(url, err)
				}
				audioCache[part.Ext] = audioPart
			}
			stream.Parts = append(stream.Parts, audioPart)
		}
		streams[itag] = stream
	}

	return &types.Data{
		Site:    "YouTube youtube.com",
		Title:   video.Title,
		Type:    "video",
		Streams: streams,
		URL:     url,
	}
}

func (e *extractor) genPartByFormat(video *youtube.Video, f *youtube.Format) (*types.Part, error) {
	ext := getStreamExt(f.MimeType)
	url, err := e.client.GetStreamURL(video, f)
	if err != nil {
		return nil, err
	}
	size := f.ContentLength
	if size == 0 {
		size, _ = request.Size(url, referer)
	}
	return &types.Part{
		URL:  url,
		Size: size,
		Ext:  ext,
	}, nil
}

func getVideoAudio(v *youtube.Video, mimeType string) (*youtube.Format, error) {
	audioFormats := v.Formats.Type(mimeType).Type("audio")
	if len(audioFormats) == 0 {
		return nil, fmt.Errorf("no audio format found after filtering")
	}
	audioFormats.Sort()
	return &audioFormats[0], nil
}

func getStreamExt(streamType string) string {
	// video/webm; codecs="vp8.0, vorbis" --> webm
	exts := utils.MatchOneOf(streamType, `(\w+)/(\w+);`)
	if exts == nil || len(exts) < 3 {
		return ""
	}
	return exts[2]
}
