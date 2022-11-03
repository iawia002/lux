# CookieMonster

[![GoDoc](https://godoc.org/github.com/MercuryEngineering/CookieMonster?status.svg)](https://godoc.org/github.com/MercuryEngineering/CookieMonster) [![Build Status](https://travis-ci.org/MercuryEngineering/CookieMonster.svg?branch=master)](https://travis-ci.org/MercuryEngineering/CookieMonster)

A simple package for parsing [Netscape Cookie File](http://curl.haxx.se/rfc/cookie_spec.html) formatted cookies into Go [Cookies](https://golang.org/pkg/net/http/#Cookie)

### Install

`go get -u github.com/MercuryEngineering/CookieMonster`

### Example

```
import (
  "fmt"
  "github.com/MercuryEngineering/CookieMonster"
)

cookies, err := cookiemonster.ParseFile("cookies.txt")
if err != nil {
  panic(err)
}

for _, cookie := range cookies {
  fmt.Printf("%s=%s\n", cookie.Name, cookie.Value)
}
```
