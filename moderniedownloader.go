package main

import (
	"fmt"
	"strings"

	scraping "github.com/hnakamur/moderniedownloader/scraping"
	vmlist "github.com/hnakamur/moderniedownloader/vmlist"
)

func main() {
	downloadPageUrl := "https://modern.ie/ja-jp/virtualization-tools#downloads"
	list, err := scraping.DownloadVmOsList(downloadPageUrl)
	if err != nil {
		panic(err)
	}

	spec := vmlist.BrowserSpec{
		OsName:       "mac",
		SoftwareName: "virtualbox",
		Version:      "8",
		OsVersion:    "win7",
	}
	files, err := vmlist.GetFilesForBrowser(strings.NewReader(list), &spec)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", files)
}
