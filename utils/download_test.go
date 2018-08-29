package utils

import (
	"reflect"
	"testing"

	"github.com/iawia002/annie/config"
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.PlaylistStart = tt.start
			config.PlaylistEnd = tt.end
			config.PlaylistItems = tt.items
			if got := NeedDownloadList(tt.args.len); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NeedDownloadList() = %v, want %v", got, tt.want)
			}
		})
	}
}
