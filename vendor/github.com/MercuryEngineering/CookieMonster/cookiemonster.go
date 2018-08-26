// Copyright (c) 2018 Henry Slawniak <https://datacenterscumbags.com/>
// Copyright (c) 2018 Mercury Engineering <https://mercury.engineering/>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Package cookiemonster provides methods for parsing Netscape format cookie files into slices of http.Cookie
package cookiemonster

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// ParseFile parses the file located at path and will return a slice of *http.Cookie or an error
func ParseFile(path string) ([]*http.Cookie, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return Parse(f)
}

// ParseString parses s and will return a slice of *http.Cookie or an error
func ParseString(s string) ([]*http.Cookie, error) {
	return Parse(bytes.NewReader([]byte(s)))
}

// Parse will parse r and will return a slice of *http.Cookie or an error
func Parse(r io.Reader) ([]*http.Cookie, error) {
	scanner := bufio.NewScanner(r)

	cookies := []*http.Cookie{}

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || line == "" {
			// Ignore comments and blank lines
			continue
		}

		split := strings.Split(line, "\t")
		if len(split) < 7 {
			// Ignore lines that are not long enough
			continue
		}

		expiresSplit := strings.Split(split[4], ".")

		expiresSec, err := strconv.Atoi(expiresSplit[0])
		if err != nil {
			return nil, err
		}

		expiresNsec := 0
		if len(expiresSplit) > 1 {
			expiresNsec, err = strconv.Atoi(expiresSplit[1])
			if err != nil {
				expiresNsec = 0
			}
		}

		cookie := &http.Cookie{
			Name:     split[5],
			Value:    split[6],
			Path:     split[2],
			Domain:   split[0],
			Expires:  time.Unix(int64(expiresSec), int64(expiresNsec)),
			Secure:   strings.ToLower(split[3]) == "true",
			HttpOnly: strings.ToLower(split[1]) == "true",
		}
		cookies = append(cookies, cookie)
	}

	return cookies, nil
}
