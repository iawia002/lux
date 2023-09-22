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
		u string
	}
	tests := []struct {
		name    string
		args    args
		want    *proto.Data
		wantErr bool
	}{
		{
			name: "bilibili",
			args: args{
				u: "https://www.bilibili.com/video/BV19C4y1f7zx/?spm_id_from=333.1007.tianma.1-1-1.click",
			},
		},
		{
			name: "douyin",
			args: args{
				u: "https://www.douyin.com/video/6967223681286278436?previous_page=main_page&tab_name=home",
			},
		},
		{
			name: "iqiyi",
			args: args{
				u: "https://www.iqiyi.com/v_19rro0jdls.html#curid=350289100_6e6601aae889d0b1004586a52027c321",
			},
		},
		{
			name: "qq",
			args: args{
				u: "https://v.qq.com/x/cover/mzc002003rpvd4j/n0046ht5pn8.html",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractors.Extract(tt.args.u)
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
