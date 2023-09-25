package cctv

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/wujiu2020/lux/extractors/proto"
)

func Test_extractor_Extract(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    *proto.Data
		wantErr bool
	}{
		{
			name: "cctv",
			args: args{
				url: "https://v.cctv.com/2023/09/20/VIDEjvDBIZp7JnCVOtC0ifS4230920.shtml?spm=C12120290324.Ps6ySUNiMIzG.0.0",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &extractor{}
			got, err := e.Extract(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractor.Extract() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else {
				b, _ := json.Marshal(got)
				fmt.Println(string(b))
			}
		})
	}
}
