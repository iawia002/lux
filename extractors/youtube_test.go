package extractors

import (
	"testing"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/test"
)

func TestYoutube(t *testing.T) {
	config.InfoOnly = true
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
				Size:    60785404,
				Quality: "hd720",
			},
		},
		{
			name: "normal test",
			args: test.Args{
				URL:     "https://youtu.be/z8eFzkfto2w",
				Title:   "Circle Of Love - Rudy Mancuso",
				Size:    27183162,
				Quality: "hd720",
			},
		},
		{
			name: "normal test",
			args: test.Args{
				URL:     "https://www.youtube.com/watch?v=ASPku-eAZYs",
				Title:   "怪獸與葛林戴華德的罪行 - HD首版電影預告大首播 (Fantastic Beasts：The Crimes of Grindelwald)",
				Size:    18330678,
				Quality: "hd720",
			},
		},
		{
			name: "playlist test",
			args: test.Args{
				URL:     "https://www.youtube.com/watch?v=x2nKigmfzLQ&list=PLfyyPldyYFapPfxZOay2AoCCaNE1Ezd4_",
				Title:   "【仔細看看】《羅根》黑白版-漫畫電影新高峰?! - 超粒方",
				Size:    108941976,
				Quality: "hd720",
			},
			playlist: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.playlist {
				// playlist mode
				config.Playlist = true
				Youtube(tt.args.URL)
				// single mode
				config.Playlist = false
				Youtube(tt.args.URL)
			} else {
				data := youtubeDownload(tt.args.URL)
				test.Check(t, tt.args, data)
			}
		})
	}
}
