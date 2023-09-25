package qq

// import (
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
// 			name: "qq",
// 			args: args{
// 				url: "https://v.qq.com/x/cover/mzc00200fjo16c3/o00474v68ks.html",
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
