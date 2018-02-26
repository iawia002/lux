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

// Domain get the domain of given URL
func Domain(url string) string {
	domainPattern := `([a-z0-9][-a-z0-9]{0,62})\.` +
		`(com\.cn|com\.hk|` +
		`cn|com|net|edu|gov|biz|org|info|pro|name|xxx|xyz|` +
		`me|top|cc|tv|tt)`
	domain := Match1(domainPattern, url)[1]
	return domain
}
