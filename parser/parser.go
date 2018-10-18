package parser

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

// GetDoc return Document object of the HTML string
func GetDoc(html string) (*goquery.Document, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// GetImages find the img with a given class name
func GetImages(
	url, html, imgClass string, urlHandler func(string) string,
) (string, []downloader.URL, error) {
	var err error
	doc, err := GetDoc(html)
	if err != nil {
		return "", nil, err
	}
	title := Title(doc)
	urls := []downloader.URL{}
	urlData := downloader.URL{}
	doc.Find(fmt.Sprintf("img[class=\"%s\"]", imgClass)).Each(
		func(i int, s *goquery.Selection) {
			urlData.URL, _ = s.Attr("src")
			if urlHandler != nil {
				// Handle URL as needed
				urlData.URL = urlHandler(urlData.URL)
			}
			urlData.Size, err = request.Size(urlData.URL, url)
			if err != nil {
				return
			}
			_, urlData.Ext, err = utils.GetNameAndExt(urlData.URL)
			if err != nil {
				return
			}
			urls = append(urls, urlData)
		},
	)
	if err != nil {
		return "", nil, err
	}
	return title, urls, nil
}

// Title get title
func Title(doc *goquery.Document) string {
	var title string
	title = strings.Replace(
		strings.TrimSpace(doc.Find("h1").First().Text()), "\n", "", -1,
	)
	if title == "" {
		// Bilibili: Some movie page got no h1 tag
		title, _ = doc.Find("meta[property=\"og:title\"]").Attr("content")
	}
	if title == "" {
		title = doc.Find("title").Text()
	}
	return utils.FileName(title)
}
