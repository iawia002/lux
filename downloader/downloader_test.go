package downloader

import (
	// "os"
	"testing"

	"github.com/iawia002/annie/config"
)

func init() {
	config.RetryTimes = 100
}

func TestDownload(t *testing.T) {
	testCases := []struct {
		name string
		data VideoData
	}{
		{
			name: "normal test",
			data: VideoData{
				Site:  "douyin",
				Title: "test",
				Type:  "video",
				Formats: map[string]FormatData{
					"default": FormatData{
						URLs: []URLData{
							{
								URL:  "https://aweme.snssdk.com/aweme/v1/playwm/?video_id=v0200f9a0000bc117isuatl67cees890&line=0",
								Size: 4927877,
								Ext:  "mp4",
							},
						},
					},
				},
			},
		},
		{
			name: "multi-format test",
			data: VideoData{
				Site:  "douyin",
				Title: "test2",
				Type:  "video",
				Formats: map[string]FormatData{
					"miaopai": FormatData{
						URLs: []URLData{
							{
								URL:  "https://txycdn.miaopai.com/stream/KwR26jUGh2ySnVjYbQiFmomNjP14LtMU3vi6sQ__.mp4?ssig=6594aa01a78e78f50c65c164d186ba9e&time_stamp=1537070910786",
								Size: 4011590,
								Ext:  "mp4",
							},
						},
						Size: 4011590,
					},
					"douyin": FormatData{
						URLs: []URLData{
							{
								URL:  "https://aweme.snssdk.com/aweme/v1/playwm/?video_id=v0200f9a0000bc117isuatl67cees890&line=0",
								Size: 4927877,
								Ext:  "mp4",
							},
						},
						Size: 4927877,
					},
				},
			},
		},
		{
			name: "google video test",
			data: VideoData{
				Site:  "google",
				Title: "google video test",
				Type:  "video",
				Formats: map[string]FormatData{
					"default": FormatData{
						URLs: []URLData{
							{
								URL:  "https://r3---sn-nx5e6nez.googlevideo.com/videoplayback?mn=sn-nx5e6nez%2Csn-a5mekne7&mm=31%2C29&key=yt6&ip=149.28.131.72&dur=320.280&keepalive=yes&pl=26&source=youtube&ms=au%2Crdu&ei=48mdW-eNL8-zz7sP_a6JqA0&id=o-AHWbYcCuEnkDYwjCFqaBJVCR--bBfLIq-LuGUcShAbHY&mv=m&mt=1537067391&sparams=aitags%2Cclen%2Cdur%2Cei%2Cgir%2Cid%2Cinitcwndbps%2Cip%2Cipbits%2Citag%2Ckeepalive%2Clmt%2Cmime%2Cmm%2Cmn%2Cms%2Cmv%2Cpl%2Crequiressl%2Csource%2Cexpire&gir=yes&ipbits=0&requiressl=yes&clen=7513224&mime=video%2Fmp4&expire=1537089092&c=WEB&fvip=4&itag=133&lmt=1513573757999118&initcwndbps=345000&aitags=133%2C134%2C135%2C136%2C160%2C242%2C243%2C244%2C247%2C278&signature=A9AF6B4F23F0967B8F2A351F2D23E61D35100CB7.CB90C59B0B6E4C126EBDA3B2C3442869799FF87F&ratebypass=yes",
								Size: 7513224,
								Ext:  "mp4",
							},
						},
						Size: 7513224,
					},
				},
			},
		},
		{
			name: "image test",
			data: VideoData{
				Site:  "bcy",
				Title: "bcy image test",
				Type:  "image",
				Formats: map[string]FormatData{
					"default": FormatData{
						URLs: []URLData{
							{
								URL:  "https://img5.bcyimg.com/coser/143767/post/c0j7x/0d713eb41a614053ac6a3b146914f6bc.jpg/w650",
								Size: 56107,
								Ext:  "jpg",
							},
							{
								URL:  "https://img9.bcyimg.com/coser/143767/post/c0j7x/d17e9b8587794d939a1363c5f715014b.jpg/w650",
								Size: 142100,
								Ext:  "jpg",
							},
						},
					},
				},
			},
		},
	}
	for _, testCase := range testCases {
		err := testCase.data.Download("")
		if err != nil {
			t.Error(err)
		}
	}
}
