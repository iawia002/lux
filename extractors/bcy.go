package extractors

import (
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

// Bcy download function
func Bcy(url string) downloader.VideoData {
	html := request.Get(url)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		log.Fatal(err)
	}
	title := strings.TrimSpace(doc.Find("h1").First().Text())
	urls := []downloader.URLData{}
	urlData := downloader.URLData{}
	doc.Find("img[class=\"detail_std detail_clickable\"]").Each(
		func(i int, s *goquery.Selection) {
			urlData.URL, _ = s.Attr("src")
			// https://img9.bcyimg.com/drawer/15294/post/1799t/1f5a87801a0711e898b12b640777720f.jpg/w650
			urlData.URL = urlData.URL[:len(urlData.URL)-5]
			urlData.Size = request.Size(urlData.URL, url)
			_, urlData.Ext = urlData.GetNameAndExt()
			urls = append(urls, urlData)
		},
	)
	data := downloader.VideoData{
		Site:  "半次元 bcy.net",
		Title: utils.FileName(title),
		Type:  "image",
		URLs:  urls,
		Size:  0,
	}
	data.Download(url)
	return data
}
