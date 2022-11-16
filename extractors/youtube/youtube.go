package youtube

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/iawia002/lia/array"
	"github.com/kkdai/youtube/v2"
	"github.com/pkg/errors"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
)

func init() {
	e := New()
	extractors.Register("youtube", e)
	extractors.Register("youtu", e) // youtu.be
}

const referer = "https://www.youtube.com"

type extractor struct {
	client *youtube.Client
}

// New returns a youtube extractor.
func New() extractors.Extractor {
	return &extractor{
		client: &youtube.Client{
			HTTPClient: &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyFromEnvironment,
				},
			},
		},
	}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.Data, error) {
	if !option.Playlist {
		video, err := e.client.GetVideo(url)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return []*extractors.Data{e.youtubeDownload(url, video)}, nil
	}

	playlist, err := e.client.GetPlaylist(url)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	needDownloadItems := utils.NeedDownloadList(option.Items, option.ItemStart, option.ItemEnd, len(playlist.Videos))
	extractedData := make([]*extractors.Data, len(needDownloadItems))
	wgp := utils.NewWaitGroupPool(option.ThreadNumber)
	dataIndex := 0
	for index, videoEntry := range playlist.Videos {
		if !array.ItemInArray(index+1, needDownloadItems) {
			continue
		}

		wgp.Add()
		go func(index int, entry *youtube.PlaylistEntry, extractedData []*extractors.Data) {
			defer wgp.Done()
			video, err := e.client.VideoFromPlaylistEntry(entry)
			if err != nil {
				return
			}
			extractedData[index] = e.youtubeDownload(url, video)
		}(dataIndex, videoEntry, extractedData)
		dataIndex++
	}
	wgp.Wait()
	return extractedData, nil
}

// youtubeDownload download function for single url
func (e *extractor) youtubeDownload(url string, video *youtube.Video) *extractors.Data {
	streams := make(map[string]*extractors.Stream, len(video.Formats))
	audioCache := make(map[string]*extractors.Part)

	for i := range video.Formats {
		f := &video.Formats[i]
		itag := strconv.Itoa(f.ItagNo)
		quality := f.MimeType
		if f.QualityLabel != "" {
			quality = fmt.Sprintf("%s %s", f.QualityLabel, f.MimeType)
		}

		part, err := e.genPartByFormat(video, f)
		if err != nil {
			return extractors.EmptyData(url, err)
		}
		stream := &extractors.Stream{
			ID:      itag,
			Parts:   []*extractors.Part{part},
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
					return extractors.EmptyData(url, err)
				}
				audioPart, err = e.genPartByFormat(video, audio)
				if err != nil {
					return extractors.EmptyData(url, err)
				}
				audioCache[part.Ext] = audioPart
			}
			stream.Parts = append(stream.Parts, audioPart)
		}
		streams[itag] = stream
	}

	return &extractors.Data{
		Site:    "YouTube youtube.com",
		Title:   video.Title,
		Type:    "video",
		Streams: streams,
		URL:     url,
	}
}

func (e *extractor) genPartByFormat(video *youtube.Video, f *youtube.Format) (*extractors.Part, error) {
	ext := getStreamExt(f.MimeType)
	url, err := e.client.GetStreamURL(video, f)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	size := f.ContentLength
	if size == 0 {
		size, _ = request.Size(url, referer)
	}
	return &extractors.Part{
		URL:  url,
		Size: size,
		Ext:  ext,
	}, nil
}

func getVideoAudio(v *youtube.Video, mimeType string) (*youtube.Format, error) {
	audioFormats := v.Formats.Type(mimeType).Type("audio")
	if len(audioFormats) == 0 {
		return nil, errors.New("no audio format found after filtering")
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
