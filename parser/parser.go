package parser

import (
	"fmt"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

// GetDoc return Document object of the HTML string
func GetDoc(html string) *goquery.Document {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		log.Fatal(err)
	}
	return doc
}

// GetImages find the img with a given class name
func GetImages(
	url, html, imgClass string, urlHandler func(string) string,
) (string, []downloader.URLData) {
	doc := GetDoc(html)
	title := strings.TrimSpace(doc.Find("h1").First().Text())
	urls := []downloader.URLData{}
	urlData := downloader.URLData{}
	doc.Find(fmt.Sprintf("img[class=\"%s\"]", imgClass)).Each(
		func(i int, s *goquery.Selection) {
			urlData.URL, _ = s.Attr("src")
			if urlHandler != nil {
				// Handle URL as needed
				urlData.URL = urlHandler(urlData.URL)
			}
			urlData.Size = request.Size(urlData.URL, url)
			_, urlData.Ext = utils.GetNameAndExt(urlData.URL)
			urls = append(urls, urlData)
		},
	)
	return title, urls
}
