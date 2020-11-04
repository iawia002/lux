package cookier

import (
	"flag"
	"strings"
	"testing"
)

var testCookie = flag.Bool("test-cookie", false, "set it to test cookie")

func TestGet(t *testing.T) {
	if !*testCookie {
		t.Skip("It's not CI friendly, so we skip it")
	}

	c := Get("https://www.bilibili.com")

	if !strings.Contains(c, "sid=") {
		t.Error("cookie should contain the sid")
	}
}
