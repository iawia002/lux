package downloader

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/fatih/color"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

// URLData data struct of single URL
type URLData struct {
	URL  string
	Size int64
	Ext  string
}

// FormatData data struct of every format
type FormatData struct {
	// [URLData: {URL, Size, Ext}, ...]
	// Some video files have multiple fragments
	// and support for downloading multiple image files at once
	URLs    []URLData
	Quality string
	Size    int64 // total size of all urls
	name    string
}

type formats []FormatData

func (f formats) Len() int           { return len(f) }
func (f formats) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f formats) Less(i, j int) bool { return f[i].Size > f[j].Size }

// VideoData data struct of video info
type VideoData struct {
	Site  string
	Title string
	Type  string
	// each format has it's own URLs and Quality
	Formats       map[string]FormatData
	sortedFormats formats
}

func progressBar(size int64) *pb.ProgressBar {
	bar := pb.New64(size).SetUnits(pb.U_BYTES).SetRefreshRate(time.Millisecond * 10)
	bar.ShowSpeed = true
	bar.ShowFinalTime = true
	bar.SetMaxWidth(1000)
	return bar
}

func (data *FormatData) calculateTotalSize() {
	var size int64
	for _, urlData := range data.URLs {
		size += urlData.Size
	}
	data.Size = size
}

// Caption download danmaku, subtitles, etc
func Caption(url, refer, fileName, ext string) {
	if !config.Caption || config.InfoOnly {
		return
	}
	fmt.Println("\nDownloading captions...")
	body := request.Get(url, refer, nil)
	filePath := utils.FilePath(fileName, ext, false)
	file, fileError := os.Create(filePath)
	if fileError != nil {
		log.Fatal(fileError)
	}
	defer file.Close()
	file.WriteString(body)
}

func writeFile(
	url string, file *os.File, headers map[string]string, bar *pb.ProgressBar,
) (int64, error) {
	res := request.Request("GET", url, nil, headers)
	defer res.Body.Close()
	writer := io.MultiWriter(file, bar)
	// Note that io.Copy reads 32kb(maximum) from input and writes them to output, then repeats.
	// So don't worry about memory.
	written, copyErr := io.Copy(writer, res.Body)
	if copyErr != nil {
		return written, fmt.Errorf("file copy error: %s", copyErr)
	}
	return written, nil
}

// Save save url file
func Save(
	urlData URLData, refer, fileName string, bar *pb.ProgressBar,
) {
	filePath := utils.FilePath(fileName, urlData.Ext, false)
	fileSize, exists := utils.FileSize(filePath)
	if bar == nil {
		bar = progressBar(urlData.Size)
		bar.Start()
	}
	// Skip segment file
	// TODO: Live video URLs will not return the size
	if exists && fileSize == urlData.Size {
		fmt.Printf("%s: file already exists, skipping\n", filePath)
		bar.Add64(fileSize)
		return
	}
	if exists && fileSize != urlData.Size {
		// files with the same name but different size
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("%s: file already exists, overwriting? [y/n]", filePath)
		overwriting, _ := reader.ReadString('\n')
		overwriting = strings.Replace(overwriting, "\n", "", -1)
		if overwriting != "y" {
			return
		}
	}
	tempFilePath := filePath + ".download"
	tempFileSize, _ := utils.FileSize(tempFilePath)
	headers := map[string]string{
		"Referer": refer,
	}
	var file *os.File
	var fileError error
	if tempFileSize > 0 {
		// range start from 0, 0-1023 means the first 1024 bytes of the file
		headers["Range"] = fmt.Sprintf("bytes=%d-", tempFileSize)
		file, fileError = os.OpenFile(tempFilePath, os.O_APPEND|os.O_WRONLY, 0644)
		bar.Add64(tempFileSize)
	} else {
		file, fileError = os.Create(tempFilePath)
	}
	if fileError != nil {
		log.Fatal(fileError)
	}
	if strings.Contains(urlData.URL, "googlevideo") {
		var start, end, chunkSize int64
		chunkSize = 10 * 1024 * 1024
		remainingSize := urlData.Size
		if tempFileSize > 0 {
			start = tempFileSize
			remainingSize -= tempFileSize
		}
		chunk := remainingSize / chunkSize
		if remainingSize%chunkSize != 0 {
			chunk++
		}
		var i int64 = 1
		for ; i <= chunk; i++ {
			end = start + chunkSize - 1
			headers["Range"] = fmt.Sprintf("bytes=%d-%d", start, end)
			temp := start
			for i := 0; ; i++ {
				written, err := writeFile(urlData.URL, file, headers, bar)
				if err == nil {
					break
				} else if i+1 >= config.RetryTimes {
					log.Fatal(err)
				}
				temp += written
				headers["Range"] = fmt.Sprintf("bytes=%d-%d", temp, end)
				time.Sleep(1 * time.Second)
			}
			start = end + 1
		}
	} else {
		temp := tempFileSize
		for i := 0; ; i++ {
			written, err := writeFile(urlData.URL, file, headers, bar)
			if err == nil {
				break
			} else if i+1 >= config.RetryTimes {
				log.Fatal(err)
			}
			temp += written
			headers["Range"] = fmt.Sprintf("bytes=%d-", temp)
			time.Sleep(1 * time.Second)
		}
	}
	// close and rename temp file at the end of this function
	defer func() {
		file.Close()
		// must close the file before rename or it will cause `The process cannot access the file because it is being used by another process.` error.
		err := os.Rename(tempFilePath, filePath)
		if err != nil {
			log.Fatal(err)
		}
	}()
}

func (data FormatData) printStream() {
	blue := color.New(color.FgBlue)
	cyan := color.New(color.FgCyan)
	blue.Println(fmt.Sprintf("     [%s]  -------------------", data.name))
	if data.Quality != "" {
		cyan.Printf("     Quality:         ")
		fmt.Println(data.Quality)
	}
	cyan.Printf("     Size:            ")
	if data.Size == 0 {
		data.calculateTotalSize()
	}
	fmt.Printf("%.2f MiB (%d Bytes)\n", float64(data.Size)/(1024*1024), data.Size)
	cyan.Printf("     # download with: ")
	if data.name == "default" {
		fmt.Println("annie \"URL\"")
	} else {
		fmt.Println("annie -f " + data.name + " \"URL\"")
	}
	fmt.Println()
}

func (v *VideoData) genSortedFormats() {
	for k, data := range v.Formats {
		if data.Size == 0 {
			data.calculateTotalSize()
		}
		data.name = k
		v.Formats[k] = data
		v.sortedFormats = append(v.sortedFormats, data)
	}
	if len(v.Formats) > 1 {
		// Only multiple formats require sorting
		sort.Sort(v.sortedFormats)
	}
	if !config.InfoOnly && !config.ExtractedData &&
		config.Format != "" && config.Format != "default" {
		return
	}
	// if Formats already have "default" format, skip follow step
	if _, ok := v.Formats["default"]; ok {
		return
	}
	bestQuality := v.sortedFormats[0].name
	v.sortedFormats[0].name = "default"
	v.Formats["default"] = v.sortedFormats[0]
	delete(v.Formats, bestQuality)
}

func (v VideoData) printInfo(format string) {
	cyan := color.New(color.FgCyan)
	fmt.Println()
	cyan.Printf(" Site:      ")
	fmt.Println(v.Site)
	cyan.Printf(" Title:     ")
	fmt.Println(v.Title)
	cyan.Printf(" Type:      ")
	fmt.Println(v.Type)
	if config.InfoOnly {
		cyan.Printf(" Streams:   ")
		fmt.Println("# All available quality")
		for _, data := range v.sortedFormats {
			data.printStream()
		}
	} else {
		cyan.Printf(" Stream:   ")
		fmt.Println()
		v.Formats[format].printStream()
	}
}

// Download download urls
func (v VideoData) Download(refer string) {
	v.genSortedFormats()
	if config.ExtractedData {
		jsonData, _ := json.MarshalIndent(v, "", "    ")
		fmt.Printf("%s\n", jsonData)
		return
	}
	var format, title string
	if config.OutputName == "" {
		title = v.Title
	} else {
		title = utils.FileName(config.OutputName)
	}
	if config.Format == "" {
		format = "default"
	} else {
		format = config.Format
	}
	data, ok := v.Formats[format]
	if !ok {
		log.Println(v)
		log.Fatal("No format named " + format)
	}
	v.printInfo(format) // if InfoOnly, this func will print all formats info
	if config.InfoOnly {
		return
	}
	// Skip the complete file that has been merged
	mergedFilePath := utils.FilePath(title, "mp4", false)
	_, mergedFileExists := utils.FileSize(mergedFilePath)
	// After the merge, the file size has changed, so we do not check whether the size matches
	if mergedFileExists {
		fmt.Printf("%s: file already exists, skipping\n", mergedFilePath)
		return
	}
	bar := progressBar(data.Size)
	bar.Start()
	if len(data.URLs) == 1 {
		// only one fragment
		Save(data.URLs[0], refer, title, bar)
		bar.Finish()
		return
	}
	wgp := utils.NewWaitGroupPool(config.ThreadNumber)
	// multiple fragments
	parts := []string{}
	for index, url := range data.URLs {
		partFileName := fmt.Sprintf("%s[%d]", title, index)
		partFilePath := utils.FilePath(partFileName, url.Ext, false)
		parts = append(parts, partFilePath)

		wgp.Add()
		go func(url URLData, refer, fileName string, bar *pb.ProgressBar) {
			defer wgp.Done()
			Save(url, refer, fileName, bar)
		}(url, refer, partFileName, bar)
	}
	wgp.Wait()
	bar.Finish()

	if v.Type != "video" {
		return
	}
	// merge
	fmt.Printf("Merging video parts into %s\n", mergedFilePath)
	var err error
	if v.Site == "YouTube youtube.com" {
		err = utils.MergeAudioAndVideo(parts, mergedFilePath)
	} else {
		err = utils.MergeToMP4(parts, mergedFilePath, title)
	}
	if err != nil {
		log.Println(err)
	}
}
