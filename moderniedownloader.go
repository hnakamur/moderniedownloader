package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/hnakamur/moderniedownloader/download"
	"github.com/hnakamur/moderniedownloader/scraping"
	"github.com/hnakamur/moderniedownloader/virtualbox"
	"github.com/hnakamur/moderniedownloader/vmlist"
)

func main() {
	flag.Parse()
	vmName := flag.Arg(0)
	if vmName == "" {
		usage()
		return
	}

	vmExists, err := virtualbox.DoesVMExist(vmName)
	if err != nil {
		panic(err)
	}

	if !vmExists {
		err = setupVM(vmName)
		if err != nil {
			panic(err)
		}
	}

	err = virtualbox.StartVM(vmName)
	if err != nil {
		panic(err)
	}
}

func setupVM(vmName string) error {
	ovaFileExists, err := download.DoesOVAFileExist(vmName)
	if err != nil {
		return err
	}

	if !ovaFileExists {
		err = downloadAndBuildOVAFile(vmName)
		if err != nil {
			return err
		}
	}

	return virtualbox.ImportAndConfigureVM(vmName)
}

func downloadAndBuildOVAFile(vmName string) error {
	list, err := scraping.DownloadVmOsList()
	if err != nil {
		return err
	}

	spec, err := virtualbox.NewVMListBrowserSpecFromVMName(vmName)
	if err != nil {
		return err
	}

	files, err := vmlist.GetFilesForBrowser(strings.NewReader(list), spec)
	if err != nil {
		return err
	}

	return download.DownloadAndBuildOVAFile(files)
}

func usage() {
	fmt.Printf("Usage: moderniedownlaoder vmName\n")
	fmt.Printf("example: moderniedownloader \"IE11 - Win8.1\"\n")
}
