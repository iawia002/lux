// +build fetch

package youtube

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func init() {
	FetchTestData()
}

// ran via go generate to fetch and update the playlist response data
func FetchTestData() {
	f, err := os.Create(testPlaylistResponseDataFile)
	exitOnError(err)
	requestURL := fmt.Sprintf(playlistFetchURL, testPlaylistID)
	resp, err := http.Get(requestURL)
	exitOnError(err)
	defer resp.Body.Close()
	n, err := io.Copy(f, resp.Body)
	exitOnError(err)
	fmt.Printf("Successfully fetched playlist %s (%d bytes)\n", testPlaylistID, n)
}

func exitOnError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
