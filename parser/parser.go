package parser

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
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
func GetImages(html, imgClass string, urlHandler func(string) string) (string, []string, error) {
	doc, err := GetDoc(html)
	if err != nil {
		return "", nil, err
	}
	title := Title(doc)
	urls := make([]string, 0)
	doc.Find(fmt.Sprintf("img[class=\"%s\"]", imgClass)).Each(
		func(i int, s *goquery.Selection) {
			url, _ := s.Attr("src")
			if urlHandler != nil {
				// Handle URL as needed
				url = urlHandler(url)
			}
			urls = append(urls, url)
		},
	)
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
	return title
}
