package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/errors"
)

func findFFmpegExecutable() string {
	ffmpegFileName := "ffmpeg"
	if runtime.GOOS == "windows" {
		ffmpegFileName = "ffmpeg.exe" // 在Windows上添加.exe擴展名
	}
	// 嘗試在當前目錄中查找ffmpeg
	matches, err := filepath.Glob("./" + ffmpegFileName)
	if err == nil && len(matches) > 0 {
		// 如果在當前目錄找到了ffmpeg，直接返回這個路徑
		return "./" + ffmpegFileName
	}

	// 返回從PATH中找到的ffmpeg路徑
	return ffmpegFileName
}

func runMergeCmd(cmd *exec.Cmd, paths []string, mergeFilePath string) error {
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return errors.Errorf("%s\n%s", err, stderr.String())
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

	return runMergeCmd(exec.Command(findFFmpegExecutable(), cmds...), paths, "")
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
		findFFmpegExecutable(), "-y", "-f", "concat", "-safe", "0",
		"-i", mergeFilePath, "-c", "copy", "-bsf:a", "aac_adtstoasc", mergedFilePath,
	)
	return runMergeCmd(cmd, paths, mergeFilePath)
}

// EmbedSubtitles embeds subtitles into the video.
func EmbedSubtitles(videoPath string, subtitlePaths []string, langs []string) error {
	tempOutput := videoPath + ".temp" + filepath.Ext(videoPath)

	cmds := []string{
		"-y",
		"-i", videoPath,
	}
	for _, subPath := range subtitlePaths {
		cmds = append(cmds, "-i", subPath)
	}

	cmds = append(cmds, "-map", "0", "-dn", "-ignore_unknown")
	cmds = append(cmds, "-c", "copy")

	ext := strings.ToLower(filepath.Ext(videoPath))
	if ext == ".mp4" {
		cmds = append(cmds, "-c:s", "mov_text")
	} else if ext == ".webm" {
		cmds = append(cmds, "-c:s", "webvtt")
	}

	// Don't copy the existing subtitles
	cmds = append(cmds, "-map", "-0:s")

	for i := range subtitlePaths {
		cmds = append(cmds, "-map", fmt.Sprintf("%d:0", i+1))
		if i < len(langs) {
			lang := langs[i]
			// Convert to ISO 639-2 (3-letter code) for MP4 compatibility
			// Simple mapping for common languages
			isoLang := "und"
			baseLang := strings.ToLower(lang)
			if strings.Contains(baseLang, "-") {
				baseLang = strings.Split(baseLang, "-")[0]
			}

			switch baseLang {
			case "zh":
				isoLang = "chi"
			case "en":
				isoLang = "eng"
			case "ja":
				isoLang = "jpn"
			case "ko":
				isoLang = "kor"
			case "es":
				isoLang = "spa"
			case "fr":
				isoLang = "fre"
			case "de":
				isoLang = "ger"
			case "ru":
				isoLang = "rus"
			default:
				if len(baseLang) == 3 {
					isoLang = baseLang
				}
			}

			cmds = append(cmds, fmt.Sprintf("-metadata:s:s:%d", i), fmt.Sprintf("language=%s", isoLang))
			cmds = append(cmds, fmt.Sprintf("-metadata:s:s:%d", i), fmt.Sprintf("handler_name=%s", lang))
			cmds = append(cmds, fmt.Sprintf("-metadata:s:s:%d", i), fmt.Sprintf("title=%s", lang))
		}
	}

	cmds = append(cmds, tempOutput)

	err := runMergeCmd(exec.Command(findFFmpegExecutable(), cmds...), []string{videoPath}, "")
	if err != nil {
		return err
	}
	return os.Rename(tempOutput, videoPath)
}
