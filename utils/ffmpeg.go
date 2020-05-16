package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

func runMergeCmd(cmd *exec.Cmd, paths []string, mergeFilePath string) error {
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%s\n%s", err, stderr.String())
	}

	if mergeFilePath != "" {
		os.Remove(mergeFilePath) // nolint
	}
	// remove parts
	for _, path := range paths {
		os.Remove(path) // nolint
	}
	return nil
}

// MergeFilesWithSameExtension merges files that have the same extension into one.
// Can also handle merging audio and video.
func MergeFilesWithSameExtension(paths []string, mergedFilePath string) error {
	cmds := []string{
		"-y",
	}
	for _, path := range paths {
		cmds = append(cmds, "-i", path)
	}
	cmds = append(cmds, "-c:v", "copy", "-c:a", "copy", mergedFilePath)
	return runMergeCmd(exec.Command("ffmpeg", cmds...), paths, "")
}

// MergeToMP4 merges video parts to an MP4 file.
func MergeToMP4(paths []string, mergedFilePath string, filename string) error {
	mergeFilePath := filename + ".txt" // merge list file should be in the current directory

	// write ffmpeg input file list
	mergeFile, _ := os.Create(mergeFilePath)
	for _, path := range paths {
		mergeFile.Write([]byte(fmt.Sprintf("file '%s'\n", path))) // nolint
	}
	mergeFile.Close() // nolint

	cmd := exec.Command(
		"ffmpeg", "-y", "-f", "concat", "-safe", "-1",
		"-i", mergeFilePath, "-c", "copy", "-bsf:a", "aac_adtstoasc", mergedFilePath,
	)
	return runMergeCmd(cmd, paths, mergeFilePath)
}
