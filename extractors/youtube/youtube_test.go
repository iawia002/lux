package youtube

import (
	"testing"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/test"
)

func TestYoutube(t *testing.T) {
	config.InfoOnly = true
	config.ThreadNumber = 9
	tests := []struct {
		name     string
		args     test.Args
		playlist bool
	}{
		{
			name: "normal test",
			args: test.Args{
				URL:     "https://www.youtube.com/watch?v=Gnbch2osEeo",
				Title:   "Multifandom Mashup 2017",
				Size:    60808458,
				Quality: `720p video/mp4; codecs="avc1.4d401f"`,
			},
		},
		{
			name: "normal test",
			args: test.Args{
				URL:     "https://youtu.be/z8eFzkfto2w",
				Title:   "Circle Of Love - Rudy Mancuso",
				Size:    32970861,
				Quality: `1080p video/webm; codecs="vp9"`,
			},
		},
		{
			name: "normal test",
			args: test.Args{
				URL:     "https://www.youtube.com/watch?v=ASPku-eAZYs",
				Title:   "怪獸與葛林戴華德的罪行 - HD首版電影預告大首播 (Fantastic Beasts：The Crimes of Grindelwald)",
				Size:    51086673,
				Quality: `1080p60 video/mp4; codecs="avc1.64002a"`,
			},
		},
		{
			name: "playlist test",
			args: test.Args{
				URL:     "https://www.youtube.com/watch?v=x2nKigmfzLQ&list=PLfyyPldyYFapPfxZOay2AoCCaNE1Ezd4_",
				Title:   "【仔細看看】《羅根》黑白版-漫畫電影新高峰?! - 超粒方",
				Size:    192310804,
				Quality: `1080p video/webm; codecs="vp9"`,
			},
			playlist: true,
		},
		{
			name: "url_encoded_fmt_stream_map test",
			args: test.Args{
				URL:     "https://youtu.be/DNaOZovrSVo",
				Title:   "QNAP Success Story - Scorptec",
				Size:    16839256,
				Quality: `hd720 video/mp4; codecs="avc1.64001F, mp4a.40.2"`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				data []downloader.Data
				err  error
			)
			if tt.playlist {
				// playlist mode
				config.Playlist = true
				_, err = Download(tt.args.URL)
				test.CheckError(t, err)
			} else {
				config.Playlist = false
				data, err = Download(tt.args.URL)
				test.CheckError(t, err)
				test.Check(t, tt.args, data[0])
			}
		})
	}
}
