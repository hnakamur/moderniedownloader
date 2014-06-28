package scraping

import (
	"fmt"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

func DownloadVmOsList(url string) (string, error) {
	doc, err := goquery.NewDocument(url)
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

	return "", fmt.Errorf("vmList not found in url=%s", url)
}
