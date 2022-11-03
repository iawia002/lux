package array

import (
	"golang.org/x/exp/constraints"
)

// ItemInArray returns true if an item is in the array.
func ItemInArray[Item comparable](item Item, items []Item) bool {
	for _, v := range items {
		if item == v {
			return true
		}
	}
	return false
}

// Min returns the smallest element in the array.
func Min[Item constraints.Ordered](items ...Item) Item {
	min := items[0]
	for i := 1; i < len(items); i++ {
		if items[i] < min {
			min = items[i]
		}
	}
	return min
}

// Max returns the largest element in the array.
func Max[Item constraints.Ordered](items ...Item) Item {
	max := items[0]
	for i := 1; i < len(items); i++ {
		if items[i] > max {
			max = items[i]
		}
	}
	return max
}
