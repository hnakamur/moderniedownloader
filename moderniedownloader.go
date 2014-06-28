package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/hnakamur/moderniedownloader/download"
	"github.com/hnakamur/moderniedownloader/scraping"
	"github.com/hnakamur/moderniedownloader/vmlist"
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

	var wg sync.WaitGroup
	wg.Add(len(files))
	for i, file := range files {
		go func(fileId int, f vmlist.ChunkFile) {
			fmt.Printf("fileId: %d\n", fileId)
			fmt.Printf("md5url: %s\n", f.Md5url)
			fmt.Printf("url: %s\n", f.Url)
			fmt.Printf("localFileName: %s\n", f.GetLocalFileName())
			fmt.Println()
			download.DownloadMd5AndFileIfMd5NotMatch(f.Md5url, f.Url, f.GetLocalFileName())
			wg.Done()
		}(i, file)
	}
	wg.Wait()
	fmt.Printf("downloaded %d files\n", len(files))
}
