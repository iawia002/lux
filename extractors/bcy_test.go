package extractors

import (
	"testing"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/test"
)

func TestBcy(t *testing.T) {
	config.InfoOnly = true
	tests := []struct {
		name string
		args test.Args
	}{
		{
			name: "normal test",
			args: test.Args{
				URL:   "https://bcy.net/illust/detail/38134/2048276",
				Title: "牛奶小姐姐，草莓味的w - 半次元-二次元爱好者社区",
			},
		},
		{
			name: "normal test",
			args: test.Args{
				URL:   "https://bcy.net/coser/detail/143767/2094010",
				Title: "phx：柠檬先行预告！牧濑红莉栖 cn：三度 - 半次元-二次元爱好者社区",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, _ := Bcy(tt.args.URL)
			test.Check(t, tt.args, data)
		})
	}
}
