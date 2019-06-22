package youtube

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/iawia002/annie/request"
)

func getDownloadURL(stream url.Values, htmlPlayerFile string) (string, error) {
	var signature string
	if s := stream.Get("s"); len(s) > 0 {
		tokens, err := getSigTokens(htmlPlayerFile)
		if err != nil {
			return "", err
		}
		signature = decipherTokens(tokens, s)
	} else {
		if sig := stream.Get("sig"); len(sig) > 0 {
			signature = sig
		}
	}
	var urlString string
	if s := stream.Get("url"); len(s) > 0 {
		urlString = s
	} else if s := stream.Get("stream"); len(s) > 0 {
		if c := stream.Get("conn"); len(c) > 0 {
			urlString = c
			if urlString[len(urlString)-1] != '/' {
				urlString += "/"
			}
		}
		urlString += s
	} else {
		return "", fmt.Errorf("couldn't extract url from format")
	}
	urlString, err := url.QueryUnescape(urlString)
	if err != nil {
		return "", err
	}
	u, err := url.Parse(urlString)
	if err != nil {
		return "", err
	}
	query := u.Query()
	query.Set("ratebypass", "yes")
	if len(signature) > 0 {
		if sp := stream.Get("sp"); sp != "" {
			query.Set(sp, signature)
		} else {
			query.Set("signature", signature)
		}
	}
	u.RawQuery = query.Encode()
	return u.String(), nil
}

// The algorithm comes from https://github.com/rylio/ytdl, it's also MIT License
// Many thanks

const (
	jsvarStr   = "[a-zA-Z_\\$][a-zA-Z_0-9]*"
	reverseStr = ":function\\(a\\)\\{" +
		"(?:return )?a\\.reverse\\(\\)" +
		"\\}"
	sliceStr = ":function\\(a,b\\)\\{" +
		"return a\\.slice\\(b\\)" +
		"\\}"
	spliceStr = ":function\\(a,b\\)\\{" +
		"a\\.splice\\(0,b\\)" +
		"\\}"
	swapStr = ":function\\(a,b\\)\\{" +
		"var c=a\\[0\\];a\\[0\\]=a\\[b(?:%a\\.length)?\\];a\\[b(?:%a\\.length)?\\]=c(?:;return a)?" +
		"\\}"
)

var actionsObjRegexp = regexp.MustCompile(fmt.Sprintf(
	"var (%s)=\\{((?:(?:%s%s|%s%s|%s%s|%s%s),?\\n?)+)\\};",
	jsvarStr, jsvarStr, reverseStr, jsvarStr, sliceStr, jsvarStr, spliceStr, jsvarStr, swapStr,
))

var actionsFuncRegexp = regexp.MustCompile(fmt.Sprintf(
	"function(?: %s)?\\(a\\)\\{"+
		"a=a\\.split\\(\"\"\\);\\s*"+
		"((?:(?:a=)?%s\\.%s\\(a,\\d+\\);)+)"+
		"return a\\.join\\(\"\"\\)"+
		"\\}", jsvarStr, jsvarStr, jsvarStr,
))

var reverseRegexp = regexp.MustCompile(fmt.Sprintf("(?m)(?:^|,)(%s)%s", jsvarStr, reverseStr))
var sliceRegexp = regexp.MustCompile(fmt.Sprintf("(?m)(?:^|,)(%s)%s", jsvarStr, sliceStr))
var spliceRegexp = regexp.MustCompile(fmt.Sprintf("(?m)(?:^|,)(%s)%s", jsvarStr, spliceStr))
var swapRegexp = regexp.MustCompile(fmt.Sprintf("(?m)(?:^|,)(%s)%s", jsvarStr, swapStr))

func getSigTokens(htmlPlayerFile string) ([]string, error) {
	u, _ := url.Parse("https://www.youtube.com/watch")
	p, err := url.Parse(htmlPlayerFile)
	if err != nil {
		return nil, err
	}

	body, err := request.Get(u.ResolveReference(p).String(), referer, nil)
	if err != nil {
		return nil, err
	}

	objResult := actionsObjRegexp.FindStringSubmatch(body)
	funcResult := actionsFuncRegexp.FindStringSubmatch(body)

	if len(objResult) < 3 || len(funcResult) < 2 {
		return nil, fmt.Errorf("error parsing signature tokens")
	}
	obj := strings.Replace(objResult[1], "$", "\\$", -1)
	objBody := strings.Replace(objResult[2], "$", "\\$", -1)
	funcBody := strings.Replace(funcResult[1], "$", "\\$", -1)

	var reverseKey, sliceKey, spliceKey, swapKey string
	var result []string

	if result = reverseRegexp.FindStringSubmatch(objBody); len(result) > 1 {
		reverseKey = strings.Replace(result[1], "$", "\\$", -1)
	}
	if result = sliceRegexp.FindStringSubmatch(objBody); len(result) > 1 {
		sliceKey = strings.Replace(result[1], "$", "\\$", -1)
	}
	if result = spliceRegexp.FindStringSubmatch(objBody); len(result) > 1 {
		spliceKey = strings.Replace(result[1], "$", "\\$", -1)
	}
	if result = swapRegexp.FindStringSubmatch(objBody); len(result) > 1 {
		swapKey = strings.Replace(result[1], "$", "\\$", -1)
	}

	keys := []string{reverseKey, sliceKey, spliceKey, swapKey}
	regex, err := regexp.Compile(fmt.Sprintf("(?:a=)?%s\\.(%s)\\(a,(\\d+)\\)", obj, strings.Join(keys, "|")))
	if err != nil {
		return nil, err
	}
	results := regex.FindAllStringSubmatch(funcBody, -1)
	var tokens []string
	for _, s := range results {
		switch s[1] {
		case swapKey:
			tokens = append(tokens, "w"+s[2])
		case reverseKey:
			tokens = append(tokens, "r")
		case sliceKey:
			tokens = append(tokens, "s"+s[2])
		case spliceKey:
			tokens = append(tokens, "p"+s[2])
		}
	}
	return tokens, nil
}

func reverseStringSlice(s []string) {
	for i, j := 0, len(s)-1; i < len(s)/2; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func decipherTokens(tokens []string, sig string) string {
	var pos int
	sigSplit := strings.Split(sig, "")
	for i, l := 0, len(tokens); i < l; i++ {
		tok := tokens[i]
		if len(tok) > 1 {
			pos, _ = strconv.Atoi(string(tok[1:]))
			pos = ^^pos
		}
		switch string(tok[0]) {
		case "r":
			reverseStringSlice(sigSplit)
		case "w":
			s := sigSplit[0]
			sigSplit[0] = sigSplit[pos]
			sigSplit[pos] = s
		case "s":
			sigSplit = sigSplit[pos:]
		case "p":
			sigSplit = sigSplit[pos:]
		}
	}
	return strings.Join(sigSplit, "")
}
