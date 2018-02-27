package fuzz

import "github.com/andybalholm/cascadia"

// Fuzz is the entrypoint used by the go-fuzz framework
func Fuzz(data []byte) int {
	sel, err := cascadia.Compile(string(data))
	if err != nil {
		if sel != nil {
			panic("sel != nil on error")
		}
		return 0
	}
	return 1
}
