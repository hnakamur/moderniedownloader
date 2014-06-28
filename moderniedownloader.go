package main

import (
	"fmt"
	"os"

	vmlist "github.com/hnakamur/moderniedownloader/vmlist"
)

func main() {
	spec := vmlist.BrowserSpec{
		OsName:       "mac",
		SoftwareName: "virtualbox",
		Version:      "8",
		OsVersion:    "win7",
	}
	files, err := vmlist.GetFilesForBrowser(os.Stdin, &spec)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", files)
}
