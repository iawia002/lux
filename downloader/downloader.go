package downloader

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/cheggaaa/pb"

	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

// VideoData data struct of video info
type VideoData struct {
	Site  string
	Title string
	URL   string
	Size  int64
	Ext   string
}

// CalculateSize get size of the url
func (data *VideoData) CalculateSize() {
	res := request.Request("GET", data.URL, nil, nil)
	defer res.Body.Close()
	s := res.Header.Get("Content-Length")
	size, _ := strconv.ParseInt(s, 10, 64)
	data.Size = size
}

func (data VideoData) printInfo() {
	fmt.Println()
	fmt.Println(" Site:   ", data.Site)
	fmt.Println("Title:   ", data.Title)
	fmt.Println(" Type:   ", data.Ext)
	fmt.Printf(" Size:    %.2f MiB (%d Bytes)\n", float64(data.Size)/(1024*1024), data.Size)
	fmt.Println()
}

// URLSave save url file
func (data VideoData) URLSave() {
	data.printInfo()
	filePath := data.Title + "." + data.Ext
	fileSize := utils.FileSize(filePath)
	if fileSize == data.Size {
		fmt.Printf("%s: file already exists, skipping\n", filePath)
		return
	}
	tempFilePath := filePath + ".download"
	tempFileSize := utils.FileSize(tempFilePath)
	var headers map[string]string
	var file *os.File
	bar := pb.New64(data.Size).SetUnits(pb.U_BYTES).SetRefreshRate(time.Millisecond * 10)
	bar.ShowSpeed = true
	bar.ShowFinalTime = true
	bar.SetMaxWidth(1000)
	if tempFileSize > 0 {
		headers = map[string]string{
			// range start from zero
			"Range": fmt.Sprintf("bytes=%d-", tempFileSize),
		}
		file, _ = os.OpenFile(tempFilePath, os.O_APPEND|os.O_WRONLY, 0644)
		bar.Set64(tempFileSize)
	} else {
		file, _ = os.Create(tempFilePath)
	}
	res := request.Request("GET", data.URL, nil, headers)
	defer res.Body.Close()
	bar.Start()
	writer := io.MultiWriter(file, bar)
	io.Copy(writer, res.Body)
	bar.Finish()
	// rename the file
	err := os.Rename(tempFilePath, filePath)
	if err != nil {
		fmt.Println(err)
		return
	}
}
