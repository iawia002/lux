package iqiyi

// import (
// 	"encoding/json"
// 	"fmt"
// 	"testing"

// 	"github.com/wujiu2020/lux/extractors/proto"
// )

// func Test_extractor_Extract(t *testing.T) {
// 	type fields struct {
// 		siteType SiteType
// 	}
// 	type args struct {
// 		url string
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    []*proto.Data
// 		wantErr bool
// 	}{
// 		{
// 			name: "iqiyi",
// 			args: args{
// 				url: "https://www.iqiyi.com/v_19rro0jdls.html#curid=350289100_6e6601aae889d0b1004586a52027c321",
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			e := &extractor{
// 				siteType: tt.fields.siteType,
// 			}
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
