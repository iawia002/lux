package parser

import (
	"testing"
)

func TestGetDoc(t *testing.T) {
	type args struct {
		html string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "normal test",
			args: args{
				html: `<html><head><title>hello</title></head><body>hello</body></html>`,
			},
			want: "hello",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := GetDoc(tt.args.html)
			title := doc.Find("title").First().Text()
			if title != tt.want {
				t.Errorf("GetDoc() = %s, want %s", title, tt.want)
			}
		})
	}
}
