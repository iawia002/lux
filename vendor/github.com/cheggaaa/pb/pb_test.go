package pb

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
)

func Test_IncrementAddsOne(t *testing.T) {
	count := 5000
	bar := New(count)
	expected := 1
	actual := bar.Increment()

	if actual != expected {
		t.Errorf("Expected {%d} was {%d}", expected, actual)
	}
}

func Test_Width(t *testing.T) {
	count := 5000
	bar := New(count)
	width := 100
	bar.SetWidth(100).Callback = func(out string) {
		if len(out) != width {
			t.Errorf("Bar width expected {%d} was {%d}", len(out), width)
		}
	}
	bar.Start()
	bar.Increment()
	bar.Finish()
}

func Test_MultipleFinish(t *testing.T) {
	bar := New(5000)
	bar.Add(2000)
	bar.Finish()
	bar.Finish()
}

func TestWriteRace(t *testing.T) {
	outBuffer := &bytes.Buffer{}
	totalCount := 20
	bar := New(totalCount)
	bar.Output = outBuffer
	bar.Start()
	var wg sync.WaitGroup
	for i := 0; i < totalCount; i++ {
		wg.Add(1)
		go func() {
			bar.Increment()
			time.Sleep(250 * time.Millisecond)
			wg.Done()
		}()
	}
	wg.Wait()
	bar.Finish()
}

func Test_Format(t *testing.T) {
	bar := New(5000).Format(strings.Join([]string{
		color.GreenString("["),
		color.New(color.BgGreen).SprintFunc()("o"),
		color.New(color.BgHiGreen).SprintFunc()("o"),
		color.New(color.BgRed).SprintFunc()("o"),
		color.GreenString("]"),
	}, "\x00"))
	w := colorable.NewColorableStdout()
	bar.Callback = func(out string) {
		w.Write([]byte(out))
	}
	bar.Add(2000)
	bar.Finish()
	bar.Finish()
}

func Test_MultiCharacter(t *testing.T) {
	bar := New(5).Format(strings.Join([]string{"[[[", "---", ">>", "....", "]]"}, "\x00"))
	bar.Start()
	for i := 0; i < 5; i++ {
		time.Sleep(500 * time.Millisecond)
		bar.Increment()
	}

	time.Sleep(500 * time.Millisecond)
	bar.Finish()
}

func Test_AutoStat(t *testing.T) {
	bar := New(5)
	bar.AutoStat = true
	bar.Start()
	time.Sleep(2 * time.Second)
	//real start work
	for i := 0; i < 5; i++ {
		time.Sleep(500 * time.Millisecond)
		bar.Increment()
	}
	//real finish work
	time.Sleep(2 * time.Second)
	bar.Finish()
}

func Test_Finish_PrintNewline(t *testing.T) {
	bar := New(5)
	buf := &bytes.Buffer{}
	bar.Output = buf
	bar.Finish()

	expected := "\n"
	actual := buf.String()
	//Finish should write newline to bar.Output
	if !strings.HasSuffix(actual, expected) {
		t.Errorf("Expected %q to have suffix %q", expected, actual)
	}
}

func Test_FinishPrint(t *testing.T) {
	bar := New(5)
	buf := &bytes.Buffer{}
	bar.Output = buf
	bar.FinishPrint("foo")

	expected := "foo\n"
	actual := buf.String()
	//FinishPrint should write to bar.Output
	if !strings.HasSuffix(actual, expected) {
		t.Errorf("Expected %q to have suffix %q", expected, actual)
	}
}
