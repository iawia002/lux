package parser

import (
	"reflect"
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
			doc, _ := GetDoc(tt.args.html)
			title := doc.Find("title").First().Text()
			if title != tt.want {
				t.Errorf("GetDoc() = %s, want %s", title, tt.want)
			}
		})
	}
}

func TestGetImages(t *testing.T) {
	tests := []struct {
		name      string
		html      string
		imgClass  string
		wantTitle string
		wantURLs  []string
	}{
		{
			name:      "fail test",
			html:      `<html><head><title>hello</title></head><body><img class="test" src="test.jpg" /><img class="test2" src="test2.jpg" /></body></html>`,
			imgClass:  "test",
			wantTitle: "hello",
			wantURLs:  []string{"test.jpg"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, urls, _ := GetImages(tt.html, tt.imgClass, nil)
			if title != tt.wantTitle {
				t.Errorf("GetImages() = %s, want %s", title, tt.wantTitle)
			}
			if !reflect.DeepEqual(urls, tt.wantURLs) {
				t.Errorf("GetImages() = %v, want %v", urls, tt.wantURLs)
			}
		})
	}
}

func TestGetTitle(t *testing.T) {
	type args struct {
		html string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "title test",
			args: args{
				html: `<html><head><title>hello</title></head><body>hello</body></html>`,
			},
			want: "hello",
		},
		{
			name: "h1 test",
			args: args{
				html: `<html><head><title>hello</title></head><body><h1> aa</h1></body></html>`,
			},
			want: "aa",
		},
		{
			name: "og:title test",
			args: args{
				html: `<html><head><meta property="og:title" content="你的名字。"></head><body>hello</body></html>`,
			},
			want: "你的名字。",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := GetDoc(tt.args.html)
			title := Title(doc)
			if title != tt.want {
				t.Errorf("Title() = %s, want %s", title, tt.want)
			}
		})
	}
}
