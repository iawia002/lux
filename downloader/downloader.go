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

// URLSize get size of the url
func (data VideoData) URLSize() int64 {
	res := request.Request("GET", data.URL, nil)
	defer res.Body.Close()
	s := res.Header.Get("Content-Length")
	size, _ := strconv.ParseInt(s, 10, 64)
	return size
}

func (data VideoData) printInfo() {
	fmt.Println()
	fmt.Println(" Site:   ", data.Site)
	fmt.Println("Title:   ", data.Title)
	fmt.Println(" Type:   ", data.Ext)
	fmt.Printf(" Size:    %.2f MiB (%d Bytes)\n", float64(data.Size)/1000000.0, data.Size)
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
	res := request.Request("GET", data.URL, nil)
	defer res.Body.Close()
	file, _ := os.Create(filePath)
	bar := pb.New(int(data.Size)).SetUnits(pb.U_BYTES).SetRefreshRate(time.Millisecond * 10)
	bar.Start()
	bar.ShowSpeed = true
	bar.ShowFinalTime = true
	bar.SetMaxWidth(1000)
	writer := io.MultiWriter(file, bar)
	io.Copy(writer, res.Body)
	bar.Finish()
}
