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
	"path/filepath"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/cheggaaa/pb/v3"

	"github.com/iawia002/annie/extractors/types"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

// Options defines options used in downloading.
type Options struct {
	InfoOnly       bool
	Stream         string
	Refer          string
	OutputPath     string
	OutputName     string
	FileNameLength int
	Caption        bool

	MultiThread  bool
	ThreadNumber int
	RetryTimes   int
	ChunkSizeMB  int
	// Aria2
	UseAria2RPC bool
	Aria2Token  string
	Aria2Method string
	Aria2Addr   string
}

// Downloader is the default downloader.
type Downloader struct {
	bar    *pb.ProgressBar
	option Options
}

func progressBar(size int64) *pb.ProgressBar {
	tmpl := `{{counters .}} {{bar . "[" "=" ">" "-" "]"}} {{speed .}} {{percent . | green}} {{rtime .}}`
	return pb.New64(size).
		Set(pb.Bytes, true).
		SetMaxWidth(1000).
		SetTemplate(pb.ProgressBarTemplate(tmpl))
}

// New returns a new Downloader implementation.
func New(option Options) *Downloader {
	downloader := &Downloader{
		option: option,
	}
	return downloader
}

// caption downloads danmaku, subtitles, etc
func (downloader *Downloader) caption(url, fileName, ext string) error {
	fmt.Println("\nDownloading captions...")

	refer := downloader.option.Refer
	if refer == "" {
		refer = url
	}
	body, err := request.GetByte(url, refer, nil)
	if err != nil {
		return err
	}
	filePath, err := utils.FilePath(fileName, ext, downloader.option.FileNameLength, downloader.option.OutputPath, true)
	if err != nil {
		return err
	}
	file, fileError := os.Create(filePath)
	if fileError != nil {
		return fileError
	}
	defer file.Close() // nolint

	if _, err = file.Write(body); err != nil {
		return err
	}
	return nil
}

func (downloader *Downloader) writeFile(url string, file *os.File, headers map[string]string) (int64, error) {
	res, err := request.Request(http.MethodGet, url, nil, headers)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close() // nolint

	barWriter := downloader.bar.NewProxyWriter(file)
	// Note that io.Copy reads 32kb(maximum) from input and writes them to output, then repeats.
	// So don't worry about memory.
	written, copyErr := io.Copy(barWriter, res.Body)
	if copyErr != nil && copyErr != io.EOF {
		return written, fmt.Errorf("file copy error: %s", copyErr)
	}
	return written, nil
}

func (downloader *Downloader) save(part *types.Part, refer, fileName string) error {
	filePath, err := utils.FilePath(fileName, part.Ext, downloader.option.FileNameLength, downloader.option.OutputPath, false)
	if err != nil {
		return err
	}
	fileSize, exists, err := utils.FileSize(filePath)
	if err != nil {
		return err
	}
	// Skip segment file
	// TODO: Live video URLs will not return the size
	if exists && fileSize == part.Size {
		downloader.bar.Add64(fileSize)
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
		downloader.bar.Add64(tempFileSize)
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
		file.Close() // nolint
		if err == nil {
			os.Rename(tempFilePath, filePath) // nolint
		}
	}()

	if downloader.option.ChunkSizeMB > 0 {
		var start, end, chunkSize int64
		chunkSize = int64(downloader.option.ChunkSizeMB) * 1024 * 1024
		remainingSize := part.Size
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
				written, err := downloader.writeFile(part.URL, file, headers)
				if err == nil {
					break
				} else if i+1 >= downloader.option.RetryTimes {
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
			written, err := downloader.writeFile(part.URL, file, headers)
			if err == nil {
				break
			} else if i+1 >= downloader.option.RetryTimes {
				return err
			}
			temp += written
			headers["Range"] = fmt.Sprintf("bytes=%d-", temp)
			time.Sleep(1 * time.Second)
		}
	}

	return nil
}

func (downloader *Downloader) multiThreadSave(dataPart *types.Part, refer, fileName string) error {
	filePath, err := utils.FilePath(fileName, dataPart.Ext, downloader.option.FileNameLength, downloader.option.OutputPath, false)
	if err != nil {
		return err
	}
	fileSize, exists, err := utils.FileSize(filePath)
	if err != nil {
		return err
	}

	// Skip segment file
	// TODO: Live video URLs will not return the size
	if exists && fileSize == dataPart.Size {
		downloader.bar.Add64(fileSize)
		return nil
	}
	tmpFilePath := filePath + ".download"
	tmpFileSize, tmpExists, err := utils.FileSize(tmpFilePath)
	if err != nil {
		return err
	}
	if tmpExists {
		if tmpFileSize == dataPart.Size {
			downloader.bar.Add64(dataPart.Size)
			return os.Rename(tmpFilePath, filePath)
		}

		if err = os.Remove(tmpFilePath); err != nil {
			return err
		}
	}

	// Scan all parts
	parts, err := readDirAllFilePart(filePath, fileName, dataPart.Ext)
	if err != nil {
		return err
	}

	var unfinishedPart []*FilePartMeta
	savedSize := int64(0)
	if len(parts) > 0 {
		lastEnd := int64(-1)
		for i, part := range parts {
			// If some parts are lost, re-insert one part.
			if part.Start-lastEnd != 1 {
				newPart := &FilePartMeta{
					Index: part.Index - 0.000001,
					Start: lastEnd + 1,
					End:   part.Start - 1,
					Cur:   lastEnd + 1,
				}
				tmp := append([]*FilePartMeta{}, parts[:i]...)
				tmp = append(tmp, newPart)
				parts = append(tmp, parts[i:]...)
				unfinishedPart = append(unfinishedPart, newPart)
			}
			// When the part has been downloaded in whole, part.Cur is equal to part.End + 1
			if part.Cur <= part.End+1 {
				savedSize += part.Cur - part.Start
				if part.Cur < part.End+1 {
					unfinishedPart = append(unfinishedPart, part)
				}
			} else {
				// The size of this part has been saved greater than the part size, delete it transparently and re-download.
				err = os.Remove(filePartPath(filePath, part))
				if err != nil {
					return err
				}
				part.Cur = part.Start
				unfinishedPart = append(unfinishedPart, part)
			}
			lastEnd = part.End
		}
		if lastEnd != dataPart.Size-1 {
			newPart := &FilePartMeta{
				Index: parts[len(parts)-1].Index + 1,
				Start: lastEnd + 1,
				End:   dataPart.Size - 1,
				Cur:   lastEnd + 1,
			}
			parts = append(parts, newPart)
			unfinishedPart = append(unfinishedPart, newPart)
		}
	} else {
		var start, end, partSize int64
		var i float32
		partSize = dataPart.Size / int64(downloader.option.ThreadNumber)
		i = 0
		for start < dataPart.Size {
			end = start + partSize - 1
			if end > dataPart.Size {
				end = dataPart.Size - 1
			} else if int(i+1) == downloader.option.ThreadNumber && end < dataPart.Size {
				end = dataPart.Size - 1
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
		downloader.bar.Add64(savedSize)
		if savedSize == dataPart.Size {
			return mergeMultiPart(filePath, parts)
		}
	}

	wgp := utils.NewWaitGroupPool(downloader.option.ThreadNumber)
	var errs []error
	for _, part := range unfinishedPart {
		wgp.Add()
		go func(part *FilePartMeta) {
			file, err := os.OpenFile(filePartPath(filePath, part), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				errs = append(errs, err)
				return
			}
			defer func() {
				file.Close() // nolint
				wgp.Done()
			}()

			var end, chunkSize int64
			headers := map[string]string{
				"Referer": refer,
			}
			if downloader.option.ChunkSizeMB <= 0 {
				chunkSize = part.End - part.Start + 1
			} else {
				chunkSize = int64(downloader.option.ChunkSizeMB) * 1024 * 1024
			}
			end = computeEnd(part.Cur, chunkSize, part.End)
			remainingSize := part.End - part.Cur + 1
			if part.Cur == part.Start {
				// Only write part to new file.
				err = writeFilePartMeta(file, part)
				if err != nil {
					errs = append(errs, err)
					return
				}
			}
			for remainingSize > 0 {
				end = computeEnd(part.Cur, chunkSize, part.End)
				headers["Range"] = fmt.Sprintf("bytes=%d-%d", part.Cur, end)
				temp := part.Cur
				for i := 0; ; i++ {
					written, err := downloader.writeFile(dataPart.URL, file, headers)
					if err == nil {
						remainingSize -= chunkSize
						break
					} else if i+1 >= downloader.option.RetryTimes {
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

func filePartPath(filepath string, part *FilePartMeta) string {
	return fmt.Sprintf("%s.part%f", filepath, part.Index)
}

func computeEnd(s, chunkSize, max int64) int64 {
	var end int64
	end = s + chunkSize - 1
	if end > max {
		end = max
	}
	return end
}

func readDirAllFilePart(filePath, filename, extname string) ([]*FilePartMeta, error) {
	dirPath := filepath.Dir(filePath)
	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}
	defer dir.Close() // nolint
	fns, err := dir.Readdir(0)
	if err != nil {
		return nil, err
	}
	var metas []*FilePartMeta
	reg := regexp.MustCompile(fmt.Sprintf("%s.%s.part.+", regexp.QuoteMeta(filename), extname))
	for _, fn := range fns {
		if reg.MatchString(fn.Name()) {
			meta, err := parseFilePartMeta(path.Join(dirPath, fn.Name()), fn.Size())
			if err != nil {
				return nil, err
			}
			metas = append(metas, meta)
		}
	}
	sort.Slice(metas, func(i, j int) bool {
		return metas[i].Index < metas[j].Index
	})
	return metas, nil
}

func parseFilePartMeta(filepath string, fileSize int64) (*FilePartMeta, error) {
	meta := new(FilePartMeta)
	size := binary.Size(*meta)
	file, err := os.OpenFile(filepath, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close() // nolint
	var buf [512]byte
	readSize, err := file.ReadAt(buf[0:size], 0)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if readSize < size {
		return nil, fmt.Errorf("the file has been broked, please delete all part files and re-download")
	}
	err = binary.Read(bytes.NewBuffer(buf[:size]), binary.LittleEndian, meta)
	if err != nil {
		return nil, err
	}
	savedSize := fileSize - int64(binary.Size(meta))
	meta.Cur = meta.Start + savedSize
	return meta, nil
}

func writeFilePartMeta(file *os.File, meta *FilePartMeta) error {
	return binary.Write(file, binary.LittleEndian, meta)
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
			f.Close()           // nolint
			os.Remove(f.Name()) // nolint
		}
	}()
	for _, part := range parts {
		file, err := os.Open(filePartPath(filepath, part))
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
	tempFile.Close() // nolint
	err = os.Rename(tempFilePath, filepath)
	return err
}

func (downloader *Downloader) aria2(title string, stream *types.Stream) error {
	rpcData := Aria2RPCData{
		JSONRPC: "2.0",
		ID:      "annie", // can be modified
		Method:  "aria2.addUri",
	}
	rpcData.Params[0] = "token:" + downloader.option.Aria2Token
	var urls []string
	for _, p := range stream.Parts {
		urls = append(urls, p.URL)
	}
	var inputs Aria2Input
	inputs.Header = append(inputs.Header, "Referer: "+downloader.option.Refer)
	for i := range urls {
		rpcData.Params[1] = urls[i : i+1]
		inputs.Out = fmt.Sprintf("%s[%d].%s", title, i, stream.Parts[0].Ext)
		rpcData.Params[2] = &inputs
		jsonData, err := json.Marshal(rpcData)
		if err != nil {
			return err
		}
		reqURL := fmt.Sprintf("%s://%s/jsonrpc", downloader.option.Aria2Method, downloader.option.Aria2Addr)
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
		res.Body.Close() // nolint
	}
	return nil
}

// Download download urls
func (downloader *Downloader) Download(data *types.Data) error {
	sortedStreams := genSortedStreams(data.Streams)
	if downloader.option.InfoOnly {
		printInfo(data, sortedStreams)
		return nil
	}

	title := downloader.option.OutputName
	if title == "" {
		title = data.Title
	}
	title = utils.FileName(title, "", downloader.option.FileNameLength)

	streamName := downloader.option.Stream
	if streamName == "" {
		streamName = sortedStreams[0].ID
	}
	stream, ok := data.Streams[streamName]
	if !ok {
		return fmt.Errorf("no stream named %s", streamName)
	}

	printStreamInfo(data, stream)

	// download caption
	if downloader.option.Caption && data.Caption != nil {
		downloader.caption(data.Caption.URL, title, data.Caption.Ext) // nolint
	}

	// Use aria2 rpc to download
	if downloader.option.UseAria2RPC {
		return downloader.aria2(title, stream)
	}

	// Skip the complete file that has been merged
	mergedFilePath, err := utils.FilePath(title, stream.Ext, downloader.option.FileNameLength, downloader.option.OutputPath, false)
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

	downloader.bar = progressBar(stream.Size)
	downloader.bar.Start()
	if len(stream.Parts) == 1 {
		// only one fragment
		var err error
		if downloader.option.MultiThread {
			err = downloader.multiThreadSave(stream.Parts[0], data.URL, title)
		} else {
			err = downloader.save(stream.Parts[0], data.URL, title)
		}

		if err != nil {
			return err
		}
		downloader.bar.Finish()
		return nil
	}

	wgp := utils.NewWaitGroupPool(downloader.option.ThreadNumber)
	// multiple fragments
	errs := make([]error, 0)
	lock := sync.Mutex{}
	parts := make([]string, len(stream.Parts))
	for index, part := range stream.Parts {
		if len(errs) > 0 {
			break
		}

		partFileName := fmt.Sprintf("%s[%d]", title, index)
		partFilePath, err := utils.FilePath(partFileName, part.Ext, downloader.option.FileNameLength, downloader.option.OutputPath, false)
		if err != nil {
			return err
		}
		parts[index] = partFilePath

		wgp.Add()
		go func(part *types.Part, fileName string) {
			defer wgp.Done()
			err := downloader.save(part, data.URL, fileName)
			if err != nil {
				lock.Lock()
				errs = append(errs, err)
				lock.Unlock()
			}
		}(part, partFileName)
	}
	wgp.Wait()
	if len(errs) > 0 {
		return errs[0]
	}
	downloader.bar.Finish()

	if data.Type != types.DataTypeVideo {
		return nil
	}

	fmt.Printf("Merging video parts into %s\n", mergedFilePath)
	if stream.Ext != "mp4" || stream.NeedMux {
		return utils.MergeFilesWithSameExtension(parts, mergedFilePath)
	}
	return utils.MergeToMP4(parts, mergedFilePath, title)
}
