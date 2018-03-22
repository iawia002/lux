# Contribution Guide

## Style Guide
### Code format
Annie uses [gofmt](https://golang.org/cmd/gofmt) to format the code, you must use [gofmt](https://golang.org/cmd/gofmt) to format your code before submitting.

### Golint
You can use [golint](https://github.com/golang/lint) to check your code format.


## Build

Make sure that this folder is in `GOPATH`, then:

```bash
$ go build
```

## Features Requested
There are several features requested by the community. If you have any idea, feel free to fork the repo, follow the style guide above, push and merge it after passing the test.

 - [ ] enable annie to download youtube playlists
 - [ ] enable annie to rename the downloaded file
 - [ ] enable annie to specifiy the file format before downloading
 - [ ] enable annie to support other sources, including facebook