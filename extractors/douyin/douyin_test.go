package douyin

// import (
// 	_ "embed"
// 	"encoding/json"
// 	"fmt"
// 	"testing"

// 	"github.com/wujiu2020/lux/extractors/proto"
// )

// func Test_extractor_Extract(t *testing.T) {
// 	type args struct {
// 		url string
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    []*proto.Data
// 		wantErr bool
// 	}{
// 		{
// 			name: "douyin",
// 			args: args{
// 				url: "https://www.douyin.com/video/6967223681286278436?previous_page=main_page&tab_name=home",
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			e := &extractor{}
// 			got, err := e.Extract(tt.args.url)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("extractor.Extract() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			b, _ := json.Marshal(got)
// 			fmt.Println(string(b))
// 			// if !reflect.DeepEqual(got, tt.want) {
// 			// 	t.Errorf("extractor.Extract() = %v, want %v", got, tt.want)
// 			// }
// 		})
// 	}
// }
