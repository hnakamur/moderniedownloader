package main

import (
	"fmt"
	"os"

	simplejson "github.com/bitly/go-simplejson"
	vmlist "github.com/hnakamur/moderniedownloader/vmlist"
)

func main() {
	osList, err := simplejson.NewFromReader(os.Stdin)
	if err != nil {
		panic(err)
	}

	spec := vmlist.BrowserSpec{
		OsName:       "mac",
		SoftwareName: "virtualbox",
		Version:      "8",
		OsVersion:    "win7",
	}
	files, err := vmlist.GetFilesForBrowser(osList, &spec)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", files)
}
