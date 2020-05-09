package utils

import (
	"os"
	"reflect"
	"testing"
)

func TestGetStringFromJSON(t *testing.T) {
	type args struct {
		json string
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "json happy path test1",
			args: args{
				json: `{"name":{"first":"Janet","last":"Prichard"},"age":47}`,
				path: "name.first",
			},
			want: "Janet",
		},
		{
			name: "json happy path test2",
			args: args{
				json: `{"children": ["Sara","Alex","Jack"]}`,
				path: "children.1",
			},
			want: "Alex",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetStringFromJSON(tt.args.json, tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStringFromJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchOneOf(t *testing.T) {
	type args struct {
		patterns []string
		text     string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "normal test",
			args: args{
				patterns: []string{`aaa(\d+)`, `hello(\d+)`},
				text:     "hello12345",
			},
			want: []string{
				"hello12345", "12345",
			},
		},
		{
			name: "normal test",
			args: args{
				patterns: []string{`aaa(\d+)`, `bbb(\d+)`},
				text:     "hello12345",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MatchOneOf(tt.args.text, tt.args.patterns...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MatchOneOf() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchAll(t *testing.T) {
	type args struct {
		pattern string
		text    string
	}
	tests := []struct {
		name string
		args args
		want [][]string
	}{
		{
			name: "normal test",
			args: args{
				pattern: `hello(\d+)`,
				text:    "hello12345hello123",
			},
			want: [][]string{
				{
					"hello12345", "12345",
				},
				{
					"hello123", "123",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MatchAll(tt.args.text, tt.args.pattern); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MatchAll() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileSize(t *testing.T) {
	type args struct {
		filePath string
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "normal test",
			args: args{
				filePath: "hello",
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _, _ := FileSize(tt.args.filePath); got != tt.want {
				t.Errorf("FileSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDomain(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "normal test",
			args: args{
				url: "http://www.aa.com",
			},
			want: "aa",
		},
		{
			name: "normal test",
			args: args{
				url: "https://aa.com",
			},
			want: "aa",
		},
		{
			name: "normal test",
			args: args{
				url: "aa.cn",
			},
			want: "aa",
		},
		{
			name: "normal test",
			args: args{
				url: "www.aa.cn",
			},
			want: "aa",
		},
		{
			name: ".com.cn test",
			args: args{
				url: "http://www.aa.com.cn",
			},
			want: "aa",
		},
		{
			name: "Universal test",
			args: args{
				url: "http://aa",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Domain(tt.args.url); got != tt.want {
				t.Errorf("Domain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLimitLength(t *testing.T) {
	type args struct {
		s      string
		length int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "normal test",
			args: args{
				s:      "你好 hello",
				length: 8,
			},
			want: "你好 hello",
		},
		{
			name: "normal test",
			args: args{
				s:      "你好 hello",
				length: 6,
			},
			want: "你好 ...",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LimitLength(tt.args.s, tt.args.length); got != tt.want {
				t.Errorf("LimitLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "normal test",
			args: args{
				name: "hello/world",
			},
			want: "hello world",
		},
		{
			name: "normal test",
			args: args{
				name: "hello:world",
			},
			want: "hello：world",
		},
		{
			name: "overly long strings test",
			args: args{
				name: "super 超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长", // length 81
			},
			want: "super 超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长超级长超级...",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FileName(tt.args.name, "", 80); got != tt.want {
				t.Errorf("FileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilePath(t *testing.T) {
	type args struct {
		name   string
		ext    string
		escape bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "normal test",
			args: args{
				name:   "hello",
				ext:    "txt",
				escape: false,
			},
			want: "hello.txt",
		},
		{
			name: "normal test",
			args: args{
				name:   "hello:world",
				ext:    "txt",
				escape: true,
			},
			want: "hello：world.txt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := FilePath(tt.args.name, tt.args.ext, 80, "", tt.args.escape); got != tt.want {
				t.Errorf("FilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestItemInSlice(t *testing.T) {
	type args struct {
		i    interface{}
		list interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "int in slice test 1",
			args: args{
				i:    1,
				list: []int{1, 2},
			},
			want: true,
		},
		{
			name: "int in slice test 2",
			args: args{
				i:    1,
				list: []int{2, 3},
			},
			want: false,
		},
		{
			name: "string test 1",
			args: args{
				i:    "hello",
				list: []string{"2", "hello"},
			},
			want: true,
		},
		{
			name: "mix test 1",
			args: args{
				i:    3,
				list: []string{"2", "3"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ItemInSlice(tt.args.i, tt.args.list); got != tt.want {
				t.Errorf("ItemInSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNameAndExt(t *testing.T) {
	type args struct {
		uri string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		{
			name: "normal test",
			args: args{
				uri: "https://img9.bcyimg.com/drawer/15294/post/1799t/1f5a87801a0711e898b12b640777720f.jpg/w650",
			},
			want:  "w650",
			want1: "jpeg",
		},
		{
			name: "normal test",
			args: args{
				uri: "https://img9.bcyimg.com/drawer/15294/post/1799t/1f5a87801a0711e898b12b640777720f.jpg",
			},
			want:  "1f5a87801a0711e898b12b640777720f",
			want1: "jpg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, _ := GetNameAndExt(tt.args.uri)
			if got != tt.want {
				t.Errorf("GetNameAndExt() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetNameAndExt() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}

	// error test
	for _, u := range []string{"https://a.com/a", "test"} {
		_, _, err := GetNameAndExt(u)
		if err == nil {
			t.Error()
		}
	}
}

func TestMd5(t *testing.T) {
	type args struct {
		text string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "normal test",
			args: args{
				text: "123456",
			},
			want: "e10adc3949ba59abbe56e057f20f883e",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Md5(tt.args.text); got != tt.want {
				t.Errorf("Md5() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrintVersion(t *testing.T) {
	PrintVersion()
}

func TestReverse(t *testing.T) {
	type args struct {
		text string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "normal test",
			args: args{
				text: "123456",
			},
			want: "654321",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Reverse(tt.args.text); got != tt.want {
				t.Errorf("Reverse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRange(t *testing.T) {
	type args struct {
		min int
		max int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "normal test",
			args: args{
				min: 1,
				max: 3,
			},
			want: []int{1, 2, 3},
		},
		{
			name: "normal test",
			args: args{
				min: 2,
				max: 2,
			},
			want: []int{2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Range(tt.args.min, tt.args.max); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Range() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLineCount(t *testing.T) {
	type args struct {
		filePath string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "negative test",
			args: args{
				filePath: "hello",
			},
			want: 0,
		}, {
			name: "positive test",
			args: args{
				filePath: "./utils_test.go",
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, _ := os.Open(tt.args.filePath)
			got, _ := FileLineCounter(file)
			file.Close()
			if got < tt.want {
				t.Errorf("Got: %v - want: %v", got, tt.want)
			}
		})
	}
}

func TestParsingFile(t *testing.T) {
	type args struct {
		filePath string
	}
	tests := []struct {
		name  string
		args  args
		start int
		end   int
		items string
		want  int
	}{
		{
			name: "negative test",
			args: args{
				filePath: "hello",
			},
			want: 0,
		}, {
			name: "start from x | end at x",
			args: args{
				filePath: "./utils_test.go",
			},
			start: 2,
			end:   4,
			want:  3,
		}, {
			name: "end at x",
			args: args{
				filePath: "./utils_test.go",
			},
			end:  4,
			want: 4,
		}, {
			name: "lower end then start",
			args: args{
				filePath: "./utils_test.go",
			},
			start: 2,
			end:   1,
			want:  1,
		}, {
			name: "items 1",
			args: args{
				filePath: "./utils_test.go",
			},
			items: "1-2, 5, 6, 8",
			want:  5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, _ := os.Open(tt.args.filePath)
			got := ParseInputFile(file, tt.items, tt.start, tt.end)
			file.Close()
			if len(got) != tt.want {
				t.Errorf("Got: %v - want: %v", len(got), tt.want)
			}
		})
	}

	// test for start from x
	t.Run("start from x", func(t *testing.T) {
		start := 5
		filePath := "./utils_test.go"
		file, _ := os.Open(filePath)
		linesCount, _ := FileLineCounter(file)
		file.Close()

		file, _ = os.Open(filePath)
		got := ParseInputFile(file, "", start, 0)
		defer file.Close()

		// start from line x to the end of the file
		// remember that the slices begin with 0 thats why it finds one line less
		if len(got) != linesCount-start {
			t.Errorf("Got: %v - want: %v", len(got), linesCount-start)
		}
	})
}
