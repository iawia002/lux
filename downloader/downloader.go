package downloader

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/cheggaaa/pb"

	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

func progressBar(size int64) *pb.ProgressBar {
	bar := pb.New64(size).SetUnits(pb.U_BYTES).SetRefreshRate(time.Millisecond * 10)
	bar.ShowSpeed = true
	bar.ShowFinalTime = true
	bar.SetMaxWidth(1000)
	return bar
}

// Caption download danmaku, subtitles, etc
func Caption(url, refer, fileName, ext string) error {
	if !config.Caption || config.InfoOnly {
		return nil
	}
	fmt.Println("\nDownloading captions...")
	body, err := request.GetByte(url, refer, nil)
	if err != nil {
		return err
	}
	filePath, err := utils.FilePath(fileName, ext, true)
	if err != nil {
		return err
	}
	file, fileError := os.Create(filePath)
	if fileError != nil {
		return fileError
	}
	defer file.Close()

	if _, err = file.Write(body); err != nil {
		return err
	}
	return nil
}

func writeFile(
	url string, file *os.File, headers map[string]string, bar *pb.ProgressBar,
) (int64, error) {
	res, err := request.Request(http.MethodGet, url, nil, headers)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	writer := io.MultiWriter(file, bar)
	// Note that io.Copy reads 32kb(maximum) from input and writes them to output, then repeats.
	// So don't worry about memory.
	written, copyErr := io.Copy(writer, res.Body)
	if copyErr != nil && copyErr != io.EOF {
		return written, fmt.Errorf("file copy error: %s", copyErr)
	}
	return written, nil
}

// Save save url file
func Save(
	urlData URL, refer, fileName string, bar *pb.ProgressBar, chunkSizeMB int,
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

	// close and rename temp file at the end of this function
	defer func() {
		// must close the file before rename or it will cause
		// `The process cannot access the file because it is being used by another process.` error.
		file.Close()
		if err == nil {
			os.Rename(tempFilePath, filePath)
		}
	}()

	if chunkSizeMB > 0 {
		var start, end, chunkSize int64
		chunkSize = int64(chunkSizeMB) * 1024 * 1024
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

	return nil
}

func MultiThreadSave(
	urlData URL, refer, fileName string, bar *pb.ProgressBar, chunkSizeMB, threadNum int,
) error {
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
	// Scan all parts
	parts, err := readDirAllFilePart(fileName, urlData.Ext)
	if err != nil {
		return err
	}

	var unfinishedPart []*FilePartMeta
	savedSize := int64(0)
	if len(parts) > 0 {
		for _, part := range parts {
			if part.Cur <= part.End {
				savedSize += part.Cur - part.Start
				if part.Cur < part.End {
					unfinishedPart = append(unfinishedPart, part)
				}
			} else if part.Cur > part.End {
				// The part saved size greater than the part size, print the error and re-download the part
				return fmt.Errorf("Part%f saved size greater than the size of this part, please delete the part and re-download\n", part.Index)
			}
		}
	} else {
		var start, end, partSize int64
		var i float32
		partSize = urlData.Size / int64(threadNum)
		i = 0
		for start < urlData.Size {
			end = start + partSize - 1
			if end > urlData.Size {
				end = urlData.Size - 1
			} else if int(i+1) == threadNum && end < urlData.Size {
				end = urlData.Size - 1
			}
			part := &FilePartMeta{
				Index: i,
				Start: start,
				End:   end,
				Cur:   start,
			}
			parts = append(parts, part)
			unfinishedPart = append(unfinishedPart, part)
			start = end + 1
			i++
		}
	}
	if savedSize > 0 {
		bar.Add64(savedSize)
		if savedSize == urlData.Size {
			return mergeMultiPart(filePath, parts)
		}
	}

	wgp := utils.NewWaitGroupPool(threadNum)
	var errs []error
	for _, part := range unfinishedPart {
		wgp.Add()
		go func(part *FilePartMeta) {
			file, err := os.OpenFile(fmt.Sprintf("%s.part%f", filePath, part.Index), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				errs = append(errs, err)
				return
			}
			defer func() {
				file.Close()
				wgp.Done()
			}()
			var end, chunkSize int64
			headers := map[string]string{
				"Referer": refer,
			}
			if chunkSizeMB <= 0 {
				chunkSize = part.End - part.Start + 1
			} else {
				chunkSize = int64(chunkSizeMB) * 1024 * 1024
			}
			end = computeEnd(part.Cur, chunkSize, part.End)
			remainingSize := part.End - part.Cur + 1
			err = writeFilePartMeta(file, part)
			if err != nil {
				errs = append(errs, err)
				return
			}
			for remainingSize > 0 {
				end = computeEnd(part.Cur, chunkSize, part.End)
				headers["Range"] = fmt.Sprintf("bytes=%d-%d", part.Cur, end)
				temp := part.Cur
				for i := 0; ; i++ {
					written, err := writeFile(urlData.URL, file, headers, bar)
					if err == nil {
						remainingSize -= chunkSize
						break
					} else if i+1 >= config.RetryTimes {
						errs = append(errs, err)
						return
					}
					temp += written
					headers["Range"] = fmt.Sprintf("bytes=%d-%d", temp, end)
				}
			}
			part.Cur = end + 1
		}(part)
	}
	wgp.Wait()
	if len(errs) > 0 {
		return errs[0]
	}
	return mergeMultiPart(filePath, parts)
}

func computeEnd(s, chunkSize, max int64) int64 {
	var end int64
	end = s + chunkSize - 1
	if end > max {
		end = max
	}
	return end
}

func parseFilePartMeta(filepath string) (*FilePartMeta, error) {
	meta := new(FilePartMeta)
	size := binary.Size(*meta)
	buf, err := mustReadFile(filepath, 0, size)
	if err != nil {
		return nil, err
	}
	err = binary.Read(bytes.NewBuffer(buf[:size]), binary.LittleEndian, meta)
	if err != nil {
		return nil, err
	}
	return meta, nil
}

func writeFilePartMeta(file *os.File, meta *FilePartMeta) error {
	return binary.Write(file, binary.LittleEndian, meta)
}

func mustReadFile(filepath string, off int64, n int) ([]byte, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var buf [512]byte
	readSize, err := file.ReadAt(buf[0:n], off)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if readSize < n {
		return nil, fmt.Errorf("There have no such size of chunk.\n")
	}
	return buf[0:n], nil
}

func readDirAllFilePart(filename, extname string) ([]*FilePartMeta, error) {
	dirPath := path.Dir(filename)
	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	fns, err := dir.Readdir(0)
	if err != nil {
		return nil, err
	}
	var metas []*FilePartMeta
	reg := regexp.MustCompile(fmt.Sprintf("%s.%s.part.+", filename, extname))

	for _, fn := range fns {
		if reg.MatchString(fn.Name()) {
			meta, err := parseFilePartMeta(fn.Name())
			if err != nil {
				return nil, err
			}
			realSize := fn.Size() - int64(binary.Size(meta))
			if meta.Cur-meta.Start != realSize {
				meta.Cur = meta.Start + realSize
			}
			metas = append(metas, meta)
		}
	}
	sort.Slice(metas, func(i, j int) bool {
		return metas[i].Index < metas[j].Index
	})
	return metas, nil
}

func mergeMultiPart(filepath string, parts []*FilePartMeta) error {
	tempFilePath := filepath + ".download"
	tempFile, err := os.OpenFile(tempFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	var partFiles []*os.File
	defer func() {
		for _, f := range partFiles {
			f.Close()
			os.Remove(f.Name())
		}
	}()
	for _, part := range parts {
		file, err := os.Open(fmt.Sprintf("%s.part%f", filepath, part.Index))
		if err != nil {
			return err
		}
		partFiles = append(partFiles, file)
		_, err = file.Seek(int64(binary.Size(part)), 0)
		if err != nil {
			return err
		}
		_, err = io.Copy(tempFile, file)
		if err != nil {
			return err
		}
	}
	tempFile.Close()
	err = os.Rename(tempFilePath, filepath)
	if err != nil {
		return err
	}
	return err
}

// Download download urls
func Download(v Data, refer string, chunkSizeMB int) error {
	v.genSortedStreams()
	if config.ExtractedData {
		jsonData, _ := json.MarshalIndent(v, "", "    ")
		fmt.Printf("%s\n", jsonData)
		return nil
	}
	var (
		title  string
		stream string
	)
	if config.OutputName == "" {
		title = utils.FileName(v.Title, "")
	} else {
		title = utils.FileName(config.OutputName, "")
	}
	if config.Stream == "" {
		stream = v.sortedStreams[0].name
	} else {
		stream = config.Stream
	}
	data, ok := v.Streams[stream]
	if !ok {
		return fmt.Errorf("no stream named %s", stream)
	}
	v.printInfo(stream) // if InfoOnly, this func will print all streams info
	if config.InfoOnly {
		return nil
	}
	// Use aria2 rpc to download
	if config.UseAria2RPC {
		rpcData := Aria2RPCData{
			JSONRPC: "2.0",
			ID:      "annie", // can be modified
			Method:  "aria2.addUri",
		}
		rpcData.Params[0] = "token:" + config.Aria2Token
		var urls []string
		for _, p := range data.URLs {
			urls = append(urls, p.URL)
		}
		var inputs Aria2Input
		inputs.Header = append(inputs.Header, "Referer: "+refer)
		for i := range urls {
			rpcData.Params[1] = urls[i : i+1]
			inputs.Out = fmt.Sprintf("%s[%d].%s", title, i, data.URLs[0].Ext)
			rpcData.Params[2] = &inputs
			jsonData, err := json.Marshal(rpcData)
			if err != nil {
				return err
			}
			reqURL := fmt.Sprintf("%s://%s/jsonrpc", config.Aria2Method, config.Aria2Addr)
			req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(jsonData))
			if err != nil {
				return err
			}
			req.Header.Set("Content-Type", "application/json")

			var client = http.Client{Timeout: 30 * time.Second}
			res, err := client.Do(req)
			if err != nil {
				return err
			}
			// The http Client and Transport guarantee that Body is always
			// non-nil, even on responses without a body or responses with
			// a zero-length body.
			res.Body.Close()
		}
		return nil
	}

	// Skip the complete file that has been merged
	var (
		mergedFilePath string
		err            error
	)
	if v.Site == "YouTube youtube.com" {
		mergedFilePath, err = utils.FilePath(title, data.URLs[0].Ext, false)
	} else {
		mergedFilePath, err = utils.FilePath(title, "mp4", false)
	}
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
		var err error
		if config.MultiThread {
			err = MultiThreadSave(data.URLs[0], refer, title, bar, chunkSizeMB, config.ThreadNumber)
		} else {
			err = Save(data.URLs[0], refer, title, bar, chunkSizeMB)
		}

		if err != nil {
			return err
		}
		bar.Finish()
		return nil
	}
	wgp := utils.NewWaitGroupPool(config.ThreadNumber)
	// multiple fragments
	errs := make([]error, 0)
	lock := sync.Mutex{}
	parts := make([]string, len(data.URLs))
	for index, url := range data.URLs {
		if len(errs) > 0 {
			break
		}

		partFileName := fmt.Sprintf("%s[%d]", title, index)
		partFilePath, err := utils.FilePath(partFileName, url.Ext, false)
		if err != nil {
			return err
		}
		parts[index] = partFilePath

		wgp.Add()
		go func(url URL, refer, fileName string, bar *pb.ProgressBar) {
			defer wgp.Done()
			err := Save(url, refer, fileName, bar, chunkSizeMB)
			if err != nil {
				lock.Lock()
				errs = append(errs, err)
				lock.Unlock()
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

	return err
}
