package utils

import (
	"strconv"
	"strings"

	"github.com/iawia002/annie/config"
)

// NeedDownloadList return the indices of playlist that need download
func NeedDownloadList(length int) []int {
	if config.PlaylistItems != "" {
		var items []int
		var index int
		temp := strings.Split(config.PlaylistItems, ",")
		for _, i := range temp {
			index, _ = strconv.Atoi(strings.TrimSpace(i))
			items = append(items, index)
		}
		return items
	}
	start := config.PlaylistStart
	end := config.PlaylistEnd
	if config.PlaylistStart < 1 {
		start = 1
	}
	if end == 0 {
		end = length
	}
	if end < start {
		end = start
	}
	return Range(start, end)
}
