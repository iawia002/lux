package eporner

import (
	"reflect"
	"testing"
)

func Test_getSrc(t *testing.T) {
	type args struct {
		html string
	}
	tests := []struct {
		name string
		args args
		want []*src
	}{
		{
			name: "T1",
			args: args{html: `<div class="clear"></div>
			<div id="downloaddiv" style="display:none;">
			<div id="hd-porn-dload">
			<div class="dloaddivcol">
			240p:<a href="/dload/baNFPbuIABZ/240/4307932-240p.mp4" >Download MP4 (240p, 131.79 MB)</a><br />
			360p:<a href="/dload/baNFPbuIABZ/360/4307932-360p.mp4" >Download MP4 (360p, 235.5 MB)</a><br />
			</div>
			</div>
			</div>`},
			want: []*src{
				{url: "/dload/baNFPbuIABZ/240/4307932-240p.mp4", quality: "240p", sizestr: "131.79 MB"},
				{url: "/dload/baNFPbuIABZ/360/4307932-360p.mp4", quality: "360p", sizestr: "235.5 MB"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSrc(tt.args.html); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSrc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSrcMeta(t *testing.T) {
	type args struct {
		text string
	}
	tests := []struct {
		name string
		args args
		want *src
	}{
		{
			name: "T2",
			args: args{text: "Download MP4 (240p, 131.79 MB)"},
			want: &src{quality: "240p", sizestr: "131.79 MB"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSrcMeta(tt.args.text); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSrcMeta() = %v, want %v", got, tt.want)
			}
		})
	}
}
