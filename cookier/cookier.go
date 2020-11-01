package cookier

import (
	"fmt"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/rod/lib/launcher"
)

// Get gets cookies from the user's Chrome or Edge automatically.
// The urls is the list of URLs for which applicable cookies will be fetched.
func Get(urls ...string) string {
	path, has := launcher.NewBrowser().LookPath()
	if !has {
		return ""
	}

	u := launcher.NewUserMode().Bin(path).MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect().DefaultDevice(devices.Clear, false)

	var page *rod.Page
	pages := browser.MustPages()
	if pages.Empty() {
		page = browser.MustPage("")
		defer page.MustClose()
	} else {
		page = pages.First()
	}

	cookies := page.MustCookies(urls...)

	list := make([]string, 0, len(cookies))
	for _, c := range cookies {
		list = append(list, fmt.Sprintf("%s=%s", c.Name, c.Value))
	}
	return strings.Join(list, "; ")
}
