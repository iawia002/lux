package bilibili

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/wujiu2020/lux/extractors"
)

func Test_extractor_Extract(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    []*extractors.Data
		wantErr bool
	}{
		{
			name: "bilibili",
			args: args{
				url: "https://www.bilibili.com/video/BV19C4y1f7zx/?spm_id_from=333.1007.tianma.1-1-1.click",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &extractor{}
			got, err := e.Extract(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractor.Extract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			b, _ := json.Marshal(got[0])
			fmt.Println(string(b))
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("extractor.Extract() = %v, want %v", got, tt.want)
			// }
		})
	}
}
