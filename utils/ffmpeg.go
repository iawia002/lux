package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)



// MergeAudioAndVideo merge audio and video
func MergeAudioAndVideo(paths []string, mergedFilePath string) error {
	cmds := []string{
		"-y",
	}
	for _, path := range paths {
		cmds = append(cmds, "-i", path)
	}
	cmds = append(
		cmds, "-c:v", "copy", "-c:a", "aac", "-strict", "experimental",
		mergedFilePath,
	)
	return runMergeCmd(exec.Command("ffmpeg", cmds...), paths, "")
}

// MergeToMP4 merge video parts to MP4
func MergeToMP4(paths []string, mergedFilePath string, filename string) error {
	//mergeFilePath := filename + ".txt" // merge list file should be in the current directory
	mergeFilePath := mergedFilePath +"_"+filename + ".txt"

	//fmt.Printf("\n mergeFilePath %s.\n\n",mergeFilePath);
	// write ffmpeg input file list
	mergeFile, _ := os.Create(mergeFilePath)
	for _, path := range paths {
		mergeFile.Write([]byte(fmt.Sprintf("file '%s'\n", path)))
	}

	//fmt.Printf("\v",paths)

	mergeFile.Close()

	cmd := exec.Command(
		"ffmpeg", "-y", "-f", "concat", "-safe", "-1",
		"-i", mergeFilePath, "-c", "copy", "-bsf:a", "aac_adtstoasc", mergedFilePath,
	)
	//fmt.Printf("\n CMD %s.\n\n",cmd);


	err := runMergeCmd(cmd, paths, mergeFilePath)

	if err == nil {
		clearPath(paths)
		return nil
	}

	errmsg := err.Error()

	if errmsg == "1001" {
		return runMergeConvertTsAfterToMp4(mergedFilePath,filename,paths)
	}
	
	return err
}

func clearPath(paths []string){
	for _, path := range paths {
		os.Remove(path)
	}
}

/**
	*	1、ffmpeg -i in.flv -c copy out.ts （把所有的flv转成ts)
	*	2、将ts放入一个文件中(比如ts.txt),好打包一起转换。
	*	3、ffmpeg -y -f concat -safe -1 -i ts.txt -c copy -bsf:a aac_adtstoasc ok.mp4
	*/
func runMergeConvertTsAfterToMp4(mergedFilePath string,filename string,paths []string) error{
	fmt.Printf("Start runMergeConvertTsAfterToMp4")

	//先把之前的删除
	os.Remove(mergedFilePath)
	mergeFilePathNew := mergedFilePath +"_"+filename + ".ts"
	mergeFileNew, _  := os.Create(mergeFilePathNew)

	pathsNew 	:= make([]string, len(paths))
	for index, path := range paths {
		new_path := path+"_"+".ts"
		//fmt.Printf("New Path %s \n",new_path)
		exec.Command("ffmpeg", "-i", path, "-c", "copy", new_path).Run()
		mergeFileNew.Write([]byte(fmt.Sprintf("file '%s'\n", new_path)))
		os.Remove(path)

		pathsNew[index]	= new_path
	}

	//fmt.Printf("\n mergeFileNew %s.\n\n",mergeFileNew);
	cmd := exec.Command(
		"ffmpeg", "-y", "-f", "concat", "-safe", "-1",
		"-i", mergeFilePathNew, "-c", "copy", "-bsf:a", "aac_adtstoasc", mergedFilePath,
	)
	//fmt.Printf("\n NEW CMD %s.\n\n",cmd)

	err := runMergeCmd(cmd, paths, mergeFilePathNew)

	//clear path
	clearPath(pathsNew)
	return err
}

func runMergeCmd(cmd *exec.Cmd, paths []string, mergeFilePath string) error {
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%s\n%s", err, stderr.String())
	}

	checkStr := stderr.String()
	checkIndex :=strings.Index(checkStr,"h264_mp4toannexb filter failed to receive output packet")
	
	if mergeFilePath != "" {
		os.Remove(mergeFilePath)
	}
	//remove parts
	// for _, path := range paths {
	// 	os.Remove(path)
	// }

	if checkIndex > -1 {
		return fmt.Errorf("%s","1001")
	}
	
	return nil;
}
