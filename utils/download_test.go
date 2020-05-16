package utils

import (
	"reflect"
	"testing"
)

func TestNeedDownloadList(t *testing.T) {
	type args struct {
		len int
	}
	tests := []struct {
		name  string
		args  args
		want  []int
		start int
		end   int
		items string
	}{
		{
			name: "start end test 1",
			args: args{
				len: 3,
			},
			start: 2,
			end:   2,
			want:  []int{2},
		},
		{
			name: "start end test 2",
			args: args{
				len: 3,
			},
			end:  2,
			want: []int{1, 2},
		},
		{
			name: "start end test 3",
			args: args{
				len: 3,
			},
			start: 2,
			end:   0,
			want:  []int{2, 3},
		},
		{
			name: "start end test 4",
			args: args{
				len: 3,
			},
			start: 2,
			end:   1,
			want:  []int{2},
		},
		{
			name: "items test",
			args: args{
				len: 3,
			},
			items: "1, 3",
			want:  []int{1, 3},
		},
		{
			name: "from to item selection 1",
			args: args{
				len: 10,
			},
			items: "1-3, 5, 7-8, 10",
			want:  []int{1, 2, 3, 5, 7, 8, 10},
		},
		{
			name: "from to item selection 2",
			args: args{
				len: 10,
			},
			items: "1,2, 4 , 5, 7-8  , 10",
			want:  []int{1, 2, 4, 5, 7, 8, 10},
		},
		{
			name: "from to item selection 3",
			args: args{
				len: 10,
			},
			items: "5-1, 2",
			want:  []int{2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NeedDownloadList(tt.items, tt.start, tt.end, tt.args.len); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NeedDownloadList() = %v, want %v", got, tt.want)
			}
		})
	}
}
