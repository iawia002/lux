package bilibili

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/iawia002/lux/extractors"
)

func listvids(fp string) []string {
	stat, err := os.Stat(fp)
	if err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(fp, 0755)
			stat, _ = os.Stat(fp)
		} else {
			panic(err)
		}
	}
	if stat == nil {
		panic("enusre fav dir failed")
	}
	if !stat.IsDir() {
		panic(fmt.Errorf(`path: '%s' is not a dir`, fp))
	}

	items, err := os.ReadDir(fp)
	if err != nil {
		panic(err)
	}

	titles := []string{}
	for _, ele := range items {
		name := ele.Name()
		ext := filepath.Ext(name)
		switch ext {
		case ".mp4", ".mkv", ".webm":
			{
				titles = append(titles, name[:len(name)-len(ext)])
				break
			}
		}
	}
	return titles
}

func favListExtractedDataFilter(uid uint64, favlTitle string, ds []*extractors.Data, opts extractors.Options) []*extractors.Data {
	favdir := filepath.Join(opts.OutputPath, fmt.Sprintf("%d_%s", uid, favlTitle))
	titles := listvids(favdir)

	favListOnDownloadDown := makeFavListOnDownloadDown(favdir)

	var nds []*extractors.Data
	for _, edata := range ds {
		if slices.Contains(titles, edata.Title) {
			fmt.Println(">>> skip downloaded video:", edata.Title)
			continue
		}
		edata.OnDownloadDone = favListOnDownloadDown
		nds = append(nds, edata)
	}
	return nds
}

func makeFavListOnDownloadDown(favdir string) func(fp string) error {
	return func(fp string) error {
		name := filepath.Base(fp)
		dest := filepath.Join(favdir, name)
		return os.Rename(fp, dest)
	}
}
