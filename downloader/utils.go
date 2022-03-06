package downloader

import (
	"fmt"
	"sort"

	"github.com/fatih/color"

	"github.com/iawia002/lux/extractors"
)

var (
	blue = color.New(color.FgBlue)
	cyan = color.New(color.FgCyan)
)

func genSortedStreams(streams map[string]*extractors.Stream) []*extractors.Stream {
	sortedStreams := make([]*extractors.Stream, 0, len(streams))
	for _, data := range streams {
		sortedStreams = append(sortedStreams, data)
	}
	if len(sortedStreams) > 1 {
		sort.SliceStable(
			sortedStreams, func(i, j int) bool { return sortedStreams[i].Size > sortedStreams[j].Size },
		)
	}
	return sortedStreams
}

func printHeader(data *extractors.Data) {
	fmt.Println()
	cyan.Printf(" Site:      ") // nolint
	fmt.Println(data.Site)
	cyan.Printf(" Title:     ") // nolint
	fmt.Println(data.Title)
	cyan.Printf(" Type:      ") // nolint
	fmt.Println(data.Type)
}

func printStream(stream *extractors.Stream) {
	blue.Println(fmt.Sprintf("     [%s]  -------------------", stream.ID)) // nolint
	if stream.Quality != "" {
		cyan.Printf("     Quality:         ") // nolint
		fmt.Println(stream.Quality)
	}
	cyan.Printf("     Size:            ") // nolint
	fmt.Printf("%.2f MiB (%d Bytes)\n", float64(stream.Size)/(1024*1024), stream.Size)
	cyan.Printf("     # download with: ") // nolint
	fmt.Printf("lux -f %s ...\n\n", stream.ID)
}

func printInfo(data *extractors.Data, sortedStreams []*extractors.Stream) {
	printHeader(data)

	cyan.Printf(" Streams:   ") // nolint
	fmt.Println("# All available quality")
	for _, stream := range sortedStreams {
		printStream(stream)
	}
}

func printStreamInfo(data *extractors.Data, stream *extractors.Stream) {
	printHeader(data)

	cyan.Printf(" Stream:   ") // nolint
	fmt.Println()
	printStream(stream)
}
