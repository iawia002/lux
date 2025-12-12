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

// ISO 639-2 language code mapping
var langToISO = map[string]string{
	"zh": "chi", "en": "eng", "ja": "jpn", "ko": "kor",
	"es": "spa", "fr": "fre", "de": "ger", "ru": "rus",
	"pt": "por", "it": "ita", "nl": "nld", "sv": "swe",
	"no": "nor", "fi": "fin", "da": "dan", "pl": "pol",
}

// toISO639 converts language code (e.g. "en-US") to ISO 639-2 (e.g. "eng")
func toISO639(lang string) string {
	base := strings.ToLower(lang)
	if i := strings.IndexAny(base, "-_"); i != -1 {
		base = base[:i]
	}
	if iso, ok := langToISO[base]; ok {
		return iso
	}
	if len(base) == 3 {
		return base
	}
	return "und"
}

// subtitleCodec returns the appropriate subtitle codec for the container format
func subtitleCodec(ext string) string {
	switch strings.ToLower(ext) {
	case ".mp4":
		return "mov_text"
	case ".webm":
		return "webvtt"
	default:
		return ""
	}
}

// EmbedSubtitles embeds subtitles into the video.
func EmbedSubtitles(videoPath string, subtitlePaths []string, langs []string) error {
	ext := filepath.Ext(videoPath)
	tempOutput := videoPath + ".temp" + ext

	// Build ffmpeg command
	cmds := []string{"-y", "-i", videoPath}
	for _, subPath := range subtitlePaths {
		cmds = append(cmds, "-i", subPath)
	}

	cmds = append(cmds,
		"-map", "0", "-dn", "-ignore_unknown",
		"-c", "copy",
	)

	if codec := subtitleCodec(ext); codec != "" {
		cmds = append(cmds, "-c:s", codec)
	}

	// Exclude existing subtitles, then map new ones
	cmds = append(cmds, "-map", "-0:s")
	for i, lang := range langs {
		if i >= len(subtitlePaths) {
			break
		}
		iso := toISO639(lang)
		cmds = append(cmds,
			"-map", fmt.Sprintf("%d:0", i+1),
			fmt.Sprintf("-metadata:s:s:%d", i), fmt.Sprintf("language=%s", iso),
			fmt.Sprintf("-metadata:s:s:%d", i), fmt.Sprintf("handler_name=%s", lang),
			fmt.Sprintf("-metadata:s:s:%d", i), fmt.Sprintf("title=%s", lang),
		)
	}

	cmds = append(cmds, tempOutput)

	if err := runMergeCmd(exec.Command(findFFmpegExecutable(), cmds...), []string{videoPath}, ""); err != nil {
		return err
	}
	return os.Rename(tempOutput, videoPath)
}
