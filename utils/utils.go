package utils

import (
	"os"
	"regexp"
)

// Match1 return result of first match
func Match1(pattern, text string) []string {
	re := regexp.MustCompile(pattern)
	value := re.FindStringSubmatch(text)
	return value
}

// FileSize return the file size of the specified path file
func FileSize(filePath string) int64 {
	file, err := os.Stat(filePath)
	if err != nil && os.IsNotExist(err) {
		return 0
	}
	return file.Size()
}
