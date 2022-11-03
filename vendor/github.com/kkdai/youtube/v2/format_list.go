package youtube

import (
	"sort"
	"strconv"
	"strings"
)

type FormatList []Format

// FindByQuality returns the first format matching Quality or QualityLabel
func (list FormatList) FindByQuality(quality string) *Format {
	for i := range list {
		if list[i].Quality == quality || list[i].QualityLabel == quality {
			return &list[i]
		}
	}
	return nil
}

// FindByItag returns the first format matching the itag number
func (list FormatList) FindByItag(itagNo int) *Format {
	for i := range list {
		if list[i].ItagNo == itagNo {
			return &list[i]
		}
	}
	return nil
}

// Type returns a new FormatList filtered by mime type of video
func (list FormatList) Type(t string) (result FormatList) {
	for i := range list {
		if strings.Contains(list[i].MimeType, t) {
			result = append(result, list[i])
		}
	}
	return result
}

// Quality returns a new FormatList filtered by quality, quality label or itag,
// but not audio quality
func (list FormatList) Quality(quality string) (result FormatList) {
	for _, f := range list {
		itag, _ := strconv.Atoi(quality)
		if itag == f.ItagNo || strings.Contains(f.Quality, quality) || strings.Contains(f.QualityLabel, quality) {
			result = append(result, f)
		}
	}
	return result
}

// AudioChannels returns a new FormatList filtered by the matching AudioChannels
func (list FormatList) AudioChannels(n int) (result FormatList) {
	for _, f := range list {
		if f.AudioChannels == n {
			result = append(result, f)
		}
	}
	return result
}

// AudioChannels returns a new FormatList filtered by the matching AudioChannels
func (list FormatList) WithAudioChannels() (result FormatList) {
	for _, f := range list {
		if f.AudioChannels > 0 {
			result = append(result, f)
		}
	}
	return result
}

// FilterQuality reduces the format list to formats matching the quality
func (v *Video) FilterQuality(quality string) {
	v.Formats = v.Formats.Quality(quality)
	v.Formats.Sort()
}

// Sort sorts all formats fields
func (list FormatList) Sort() {
	sort.SliceStable(list, func(i, j int) bool {
		return sortFormat(i, j, list)
	})
}

// sortFormat sorts video by resolution, FPS, codec (av01, vp9, avc1), bitrate
// sorts audio by codec (mp4, opus), channels, bitrate, sample rate
func sortFormat(i int, j int, formats FormatList) bool {

	// Sort by Width
	if formats[i].Width == formats[j].Width {
		// Format 137 downloads slowly, give it less priority
		// see https://github.com/kkdai/youtube/pull/171
		switch 137 {
		case formats[i].ItagNo:
			return false
		case formats[j].ItagNo:
			return true
		}

		// Sort by FPS
		if formats[i].FPS == formats[j].FPS {
			if formats[i].FPS == 0 && formats[i].AudioChannels > 0 && formats[j].AudioChannels > 0 {
				// Audio
				// Sort by codec
				codec := map[int]int{}
				for _, index := range []int{i, j} {
					if strings.Contains(formats[index].MimeType, "mp4") {
						codec[index] = 1
					} else if strings.Contains(formats[index].MimeType, "opus") {
						codec[index] = 2
					}
				}
				if codec[i] == codec[j] {
					// Sort by Audio Channel
					if formats[i].AudioChannels == formats[j].AudioChannels {
						// Sort by Audio Bitrate
						if formats[i].Bitrate == formats[j].Bitrate {
							// Sort by Audio Sample Rate
							return formats[i].AudioSampleRate > formats[j].AudioSampleRate
						}
						return formats[i].Bitrate > formats[j].Bitrate
					}
					return formats[i].AudioChannels > formats[j].AudioChannels
				}
				return codec[i] < codec[j]
			}
			// Video
			// Sort by codec
			codec := map[int]int{}
			for _, index := range []int{i, j} {
				if strings.Contains(formats[index].MimeType, "av01") {
					codec[index] = 1
				} else if strings.Contains(formats[index].MimeType, "vp9") {
					codec[index] = 2
				} else if strings.Contains(formats[index].MimeType, "avc1") {
					codec[index] = 3
				}
			}
			if codec[i] == codec[j] {
				// Sort by Audio Bitrate
				return formats[i].Bitrate > formats[j].Bitrate
			}
			return codec[i] < codec[j]
		}
		return formats[i].FPS > formats[j].FPS
	}
	return formats[i].Width > formats[j].Width
}
