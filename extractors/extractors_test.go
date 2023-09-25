package extractors_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/wujiu2020/lux/extractors"
	"github.com/wujiu2020/lux/extractors/proto"
)

func TestExtract(t *testing.T) {
	type args struct {
		u       string
		quality string
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
				u:       "https://v.cctv.com/2023/09/21/VIDENqWdChgLFvbSOVeC1szE230921.shtml?spm=C90324.PE6LRxWJhH5P.S23920.38",
				quality: "270P",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractors.Extract(tt.args.u, tt.args.quality)
			if (err != nil) != tt.wantErr {
				t.Errorf("Extract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			b, _ := json.Marshal(got)
			fmt.Println(string(b))
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("Extract() = %v, want %v", got, tt.want)
			// }
		})
	}
}
