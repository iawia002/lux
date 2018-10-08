package downloader

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"
	"net/http"
	"bytes"

	"github.com/cheggaaa/pb"
	"github.com/fatih/color"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

// URLData data struct of single URL
type URLData struct {
	URL  string `json:"url"`
	Size int64  `json:"size"`
	Ext  string `json:"ext"`
}

// FormatData data struct of every format
type FormatData struct {
	// [URLData: {URL, Size, Ext}, ...]
	// Some video files have multiple fragments
	// and support for downloading multiple image files at once
	URLs    []URLData `json:"urls"`
	Quality string    `json:"quality"`
	// total size of all urls
	Size int64 `json:"size"`
	name string
}

// JsonRPC 2.0 for Aria2
type Aria2RPCData struct {
	// More info about RPC interface please refer to 
	// https://aria2.github.io/manual/en/html/aria2c.html#rpc-interface
	Jsonrpc string `json:"jsonrpc"`
	Id string `json:"id"`
	// For a simple download, only inplemented `addUri`
	Method string `json:"method"`
	// secret, uris, options
	Params [3]interface{} `json:"params"`
}

// Options for `aria2.addUri`
// https://aria2.github.io/manual/en/html/aria2c.html#id3
type Aria2Input struct {
	// For a simple download, only add headers
	Header []string `json:"header"`
}

type formats []FormatData

func (f formats) Len() int           { return len(f) }
func (f formats) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f formats) Less(i, j int) bool { return f[i].Size > f[j].Size }

// VideoData data struct of video info
type VideoData struct {
	Site  string `json:"site"`
	Title string `json:"title"`
	Type  string `json:"type"`
	// each format has it's own URLs and Quality
	Formats       map[string]FormatData `json:"formats"`
	sortedFormats formats
}

// EmptyData empty VideoData list
var EmptyData = make([]VideoData, 0)

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
func Caption(url, refer, fileName, ext string) error {
	if !config.Caption || config.InfoOnly {
		return nil
	}
	fmt.Println("\nDownloading captions...")
	body, err := request.Get(url, refer, nil)
	if err != nil {
		return err
	}
	filePath, err := utils.FilePath(fileName, ext, false)
	if err != nil {
		return err
	}
	file, fileError := os.Create(filePath)
	if fileError != nil {
		return fileError
	}
	defer file.Close()
	file.WriteString(body)
	return nil
}

func writeFile(
	url string, file *os.File, headers map[string]string, bar *pb.ProgressBar,
) (int64, error) {
	res, err := request.Request("GET", url, nil, headers)
	if err != nil {
		return 0, err
	}
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
) error {
	var err error
	filePath, err := utils.FilePath(fileName, urlData.Ext, false)
	if err != nil {
		return err
	}
	fileSize, exists, err := utils.FileSize(filePath)
	if err != nil {
		return err
	}
	if bar == nil {
		bar = progressBar(urlData.Size)
		bar.Start()
	}
	// Skip segment file
	// TODO: Live video URLs will not return the size
	if exists && fileSize == urlData.Size {
		bar.Add64(fileSize)
		return nil
	}
	tempFilePath := filePath + ".download"
	tempFileSize, _, err := utils.FileSize(tempFilePath)
	if err != nil {
		return err
	}
	headers := map[string]string{
		"Referer": refer,
	}
	var (
		file      *os.File
		fileError error
	)
	if tempFileSize > 0 {
		// range start from 0, 0-1023 means the first 1024 bytes of the file
		headers["Range"] = fmt.Sprintf("bytes=%d-", tempFileSize)
		file, fileError = os.OpenFile(tempFilePath, os.O_APPEND|os.O_WRONLY, 0644)
		bar.Add64(tempFileSize)
	} else {
		file, fileError = os.Create(tempFilePath)
	}
	if fileError != nil {
		return fileError
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
					return err
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
				return err
			}
			temp += written
			headers["Range"] = fmt.Sprintf("bytes=%d-", temp)
			time.Sleep(1 * time.Second)
		}
	}

	// close and rename temp file at the end of this function
	defer func() {
		// must close the file before rename or it will cause `The process cannot access the file because it is being used by another process.` error.
		file.Close()
		if err == nil {
			os.Rename(tempFilePath, filePath)
		}
	}()
	return nil
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
	fmt.Printf("annie -f %s ...\n\n", data.name)
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
		sort.Sort(v.sortedFormats)
	}
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
func (v VideoData) Download(refer string) error {
	v.genSortedFormats()
	if config.ExtractedData {
		jsonData, _ := json.MarshalIndent(v, "", "    ")
		fmt.Printf("%s\n", jsonData)
		return nil
	}
	// User aria2 rpc to download
	if config.UseAria2RPC {
		rpcData := Aria2RPCData {
			Jsonrpc: "2.0",
			Id: "annie",  // can be modified
			Method: "aria2.addUri",
		}
		rpcData.Params[0] = "token:" + config.Aria2Token
		var urls []string
		for _, p := range v.sortedFormats[0].URLs {
			urls = append(urls, p.URL)
		}
		rpcData.Params[1] = &urls
		var inputs Aria2Input
		inputs.Header = append(inputs.Header, "Referer: " + refer)
		rpcData.Params[2] = &inputs
		jsonData, _ := json.Marshal(rpcData)
		reqUrl := config.Aria2Method + "://" + config.Aria2Addr + "/jsonrpc"
		req, _ := http.NewRequest("POST", reqUrl, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		var client http.Client
		resp, _ := client.Do(req)
		defer resp.Body.Close()
		return nil
	}
	var (
		title  string
		format string
	)
	if config.OutputName == "" {
		title = v.Title
	} else {
		title = utils.FileName(config.OutputName)
	}
	if config.Format == "" {
		format = v.sortedFormats[0].name
	} else {
		format = config.Format
	}
	data, ok := v.Formats[format]
	if !ok {
		return fmt.Errorf("no format named %s", format)
	}
	v.printInfo(format) // if InfoOnly, this func will print all formats info
	if config.InfoOnly {
		return nil
	}
	var err error
	// Skip the complete file that has been merged
	mergedFilePath, err := utils.FilePath(title, "mp4", false)
	if err != nil {
		return err
	}
	_, mergedFileExists, err := utils.FileSize(mergedFilePath)
	if err != nil {
		return err
	}
	// After the merge, the file size has changed, so we do not check whether the size matches
	if mergedFileExists {
		fmt.Printf("%s: file already exists, skipping\n", mergedFilePath)
		return nil
	}
	bar := progressBar(data.Size)
	bar.Start()
	if len(data.URLs) == 1 {
		// only one fragment
		err := Save(data.URLs[0], refer, title, bar)
		if err != nil {
			return err
		}
		bar.Finish()
		return nil
	}
	wgp := utils.NewWaitGroupPool(config.ThreadNumber)
	// multiple fragments
	errs := make([]error, 0)
	parts := make([]string, len(data.URLs))
	for index, url := range data.URLs {
		partFileName := fmt.Sprintf("%s[%d]", title, index)
		partFilePath, err := utils.FilePath(partFileName, url.Ext, false)
		if err != nil {
			return err
		}
		parts = append(parts, partFilePath)

		wgp.Add()
		go func(url URLData, refer, fileName string, bar *pb.ProgressBar) {
			defer wgp.Done()
			err := Save(url, refer, fileName, bar)
			if err != nil {
				errs = append(errs, err)
			}
		}(url, refer, partFileName, bar)
	}
	wgp.Wait()
	if len(errs) > 0 {
		return errs[0]
	}
	bar.Finish()

	if v.Type != "video" {
		return nil
	}
	// merge
	fmt.Printf("Merging video parts into %s\n", mergedFilePath)
	if v.Site == "YouTube youtube.com" {
		err = utils.MergeAudioAndVideo(parts, mergedFilePath)
	} else {
		err = utils.MergeToMP4(parts, mergedFilePath, title)
	}
	if err != nil {
		return err
	}
	return nil
}
