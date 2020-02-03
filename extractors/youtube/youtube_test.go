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
				URL:   "https://www.youtube.com/watch?v=Gnbch2osEeo",
				Title: "Multifandom Mashup 2017",
			},
		},
		{
			name: "normal test",
			args: test.Args{
				URL:   "https://youtu.be/z8eFzkfto2w",
				Title: "Circle Of Love | Rudy Mancuso",
			},
		},
		{
			name: "signature test",
			args: test.Args{
				URL:   "https://www.youtube.com/watch?v=ZtgzKBrU1GY",
				Title: "Halo Infinite - E3 2019 - Discover Hope",
			},
		},
		{
			name: "normal test",
			args: test.Args{
				URL:   "https://www.youtube.com/watch?v=ASPku-eAZYs",
				Title: "怪獸與葛林戴華德的罪行 | HD首版電影預告大首播 (Fantastic Beasts: The Crimes of Grindelwald)",
			},
		},
		{
			name: "playlist test",
			args: test.Args{
				URL:   "https://www.youtube.com/watch?v=Lt2pwLxJxgA&list=PLIYAO-qLriEtYm7UcXPH3SOJxgqjwRrIw",
				Title: "papi酱 - 你有酱婶儿的朋友吗？",
			},
			playlist: true,
		},
		{
			name: "url_encoded_fmt_stream_map test",
			args: test.Args{
				URL:   "https://youtu.be/DNaOZovrSVo",
				Title: "QNAP Case Study - Scorptec",
			},
		},
		{
			name: "stream 404 test 1",
			args: test.Args{
				URL:   "https://www.youtube.com/watch?v=MRJ8NnUXacY",
				Title: "FreeFileSync: Mirror Synchronization",
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
				_, err = Extract(tt.args.URL)
				test.CheckError(t, err)
			} else {
				config.Playlist = false
				data, err = Extract(tt.args.URL)
				test.CheckError(t, err)
				test.Check(t, tt.args, data[0])
			}
		})
	}
}
