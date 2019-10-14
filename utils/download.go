package utils

import (
	"strconv"
	"strings"
)

// NeedDownloadList return the indices of playlist that need download
func NeedDownloadList(length int) []int {
	if config.items != "" {
		var items []int
		var selStart, selEnd int
		temp := strings.Split(config.items, ",")

		for _, i := range temp {
			selection := strings.Split(i, "-")
			selStart, _ = strconv.Atoi(strings.TrimSpace(selection[0]))

			if len(selection) >= 2 {
				selEnd, _ = strconv.Atoi(strings.TrimSpace(selection[1]))
			} else {
				selEnd = selStart
			}

			for item := selStart; item <= selEnd; item++ {
				items = append(items, item)
			}
		}
		return items
	}
	start := config.itemStart
	end := config.itemEnd
	if config.itemStart < 1 {
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
