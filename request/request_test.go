package request

import (
	"testing"
)

func TestGet(t *testing.T) {
	var err error
	type args struct {
		url     string
		refer   string
		headers map[string]string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "normal test",
			args: args{
				url:     "https://google.com",
				refer:   "",
				headers: nil,
			},
		},
		{
			name: "test with refer and headers",
			args: args{
				url:   "https://google.com",
				refer: "https://google.com",
				headers: map[string]string{
					"Referer": "https://google.com",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err = Get(tt.args.url, tt.args.refer, tt.args.headers)
			if err != nil {
				t.Error()
			}
		})
	}

	// error test
	_, err = Get("test", "", nil)
	if err == nil {
		t.Error()
	}

	// with config
	debug = true
	rawCookie = "name: value;"
	_, err = Get("https://google.com", "", nil)
	if err != nil {
		t.Error()
	}
}

func TestHeaders(t *testing.T) {
	var err error
	type args struct {
		url   string
		refer string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "normal test",
			args: args{
				url:   "https://google.com",
				refer: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err = Headers(tt.args.url, tt.args.refer)
			if err != nil {
				t.Error()
			}
		})
	}
}

func TestSize(t *testing.T) {
	var err error
	type args struct {
		url   string
		refer string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "normal test",
			args: args{
				url:   "https://www.google.com/images/branding/googlelogo/1x/googlelogo_color_272x92dp.png",
				refer: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err = Size(tt.args.url, tt.args.refer)
			if err != nil {
				t.Error()
			}
		})
	}
}

func TestContentType(t *testing.T) {
	type args struct {
		url   string
		refer string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "normal test",
			args: args{
				url:   "https://google.com",
				refer: "",
			},
			want: "text/html",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contentType, _ := ContentType(tt.args.url, tt.args.refer)
			if contentType != tt.want {
				t.Errorf("ContentType() = %s, want %s", contentType, tt.want)
			}
		})
	}
}
