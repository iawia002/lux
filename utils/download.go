package utils

import (
	"strconv"
	"strings"

	"github.com/iawia002/annie/config"
)

// NeedDownloadList return the indices of playlist that need download
func NeedDownloadList(length int) []int {
	if config.Items != "" {
		var items []int
		var selStart, selEnd int
		temp := strings.Split(config.Items, ",")

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
	start := config.ItemStart
	end := config.ItemEnd
	if config.ItemStart < 1 {
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
