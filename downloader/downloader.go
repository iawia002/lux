package downloader

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/cheggaaa/pb"

	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

// URLData data struct of single URL
type URLData struct {
	URL  string
	Size int64
}

// VideoData data struct of video info
type VideoData struct {
	Site  string
	Title string
	URLs  []URLData
	Size  int64
	Ext   string
}

func (data VideoData) printInfo() {
	fmt.Println()
	fmt.Println(" Site:   ", data.Site)
	fmt.Println("Title:   ", data.Title)
	fmt.Println(" Type:   ", data.Ext)
	fmt.Printf(" Size:    %.2f MiB (%d Bytes)\n", float64(data.Size)/(1024*1024), data.Size)
	fmt.Println()
}

// urlSave save url file
func (data VideoData) urlSave(
	urlData URLData, refer, fileName string, bar *pb.ProgressBar, parts *[]string,
) {
	filePath := fileName + "." + data.Ext
	fileSize := utils.FileSize(filePath)
	if parts != nil {
		*parts = append(*parts, filePath)
	}
	if fileSize == urlData.Size {
		fmt.Printf("%s: file already exists, skipping\n", filePath)
		bar.Add64(fileSize)
		return
	}
	tempFilePath := filePath + ".download"
	tempFileSize := utils.FileSize(tempFilePath)
	headers := map[string]string{
		"Referer": refer,
	}
	var file *os.File
	if tempFileSize > 0 {
		// range start from zero
		headers["Range"] = fmt.Sprintf("bytes=%d-", tempFileSize)
		file, _ = os.OpenFile(tempFilePath, os.O_APPEND|os.O_WRONLY, 0644)
		bar.Add64(tempFileSize)
	} else {
		file, _ = os.Create(tempFilePath)
	}
	res := request.Request("GET", urlData.URL, nil, headers)
	defer res.Body.Close()
	writer := io.MultiWriter(file, bar)
	io.Copy(writer, res.Body)
	// rename the file
	err := os.Rename(tempFilePath, filePath)
	if err != nil {
		fmt.Println(err)
		return
	}
}

// Download download urls
func (data VideoData) Download(refer string) {
	data.printInfo()
	bar := pb.New64(data.Size).SetUnits(pb.U_BYTES).SetRefreshRate(time.Millisecond * 10)
	bar.ShowSpeed = true
	bar.ShowFinalTime = true
	bar.SetMaxWidth(1000)
	bar.Start()
	if len(data.URLs) == 1 {
		// only one fragment
		data.urlSave(data.URLs[0], refer, data.Title, bar, nil)
		bar.Finish()
	} else {
		// multiple fragments
		parts := []string{}
		for index, url := range data.URLs {
			data.urlSave(url, refer, fmt.Sprintf("%s[%d]", data.Title, index), bar, &parts)
		}
		bar.Finish()

		// merge
		// write ffmpeg input file list
		mergeFile := data.Title + "-merge.txt"
		file, _ := os.Create(mergeFile)
		for _, part := range parts {
			file.Write([]byte(fmt.Sprintf("file '%s'\n", part)))
		}

		filePath := data.Title + "." + data.Ext
		fmt.Printf("Merging video parts into %s\n", filePath)
		cmd := exec.Command(
			"ffmpeg", "-y", "-f", "concat", "-safe", "-1",
			"-i", mergeFile, "-c", "copy", "-bsf:a", "aac_adtstoasc", filePath,
		)
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
		// remove parts
		os.Remove(mergeFile)
		for _, part := range parts {
			os.Remove(part)
		}
	}
}
