package scraping

import (
	"fmt"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

const (
	DownloadPageUrl = "https://modern.ie/ja-jp/virtualization-tools#downloads"
)

func DownloadVmOsList() (string, error) {
	doc, err := goquery.NewDocument(DownloadPageUrl)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`d\.osList=(\[.*\]);`)
	list := ""
	doc.Find("script").Each(func(_ int, s *goquery.Selection) {
		values := re.FindAllStringSubmatch(s.Text(), 1)
		if values != nil {
			list = values[0][1]
		}
	})
	if list != "" {
		return list, nil
	}

	return "", fmt.Errorf("vmList not found in url=%s", DownloadPageUrl)
}
