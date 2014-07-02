package main

import (
	"flag"
	"fmt"

	"github.com/hnakamur/moderniedownloader/download"
	"github.com/hnakamur/moderniedownloader/virtualbox"
	"github.com/hnakamur/moderniedownloader/vmlist"
)

var lflag bool
var hflag bool

func init() {
	flag.BoolVar(&lflag, "l", false, "list available modern.IE VM names")
	flag.BoolVar(&hflag, "h", false, "help")
}

func main() {
	flag.Parse()

	if lflag {
		listAvailableVmNames()
		return
	}
	if hflag || flag.NArg() == 0 {
		usage()
		return
	}

	vmName := flag.Arg(0)

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
	spec, err := virtualbox.NewVMListBrowserSpecFromVMName(vmName)
	if err != nil {
		return err
	}

	files, err := vmlist.GetFilesForBrowser(spec)
	if err != nil {
		return err
	}

	return download.DownloadAndBuildOVAFile(files)
}

func listAvailableVmNames() {
	vmNames, err := virtualbox.GetVmNameList()
	if err != nil {
		panic(err)
	}
	for _, vmName := range vmNames {
		fmt.Printf("%s\n", vmName)
	}
}

func usage() {
	fmt.Printf("Usage: moderniedownlaoder vmName\n")
	fmt.Printf("           example: moderniedownloader \"IE11 - Win8.1\"\n")
	fmt.Printf("       moderniedownlaoder -l     list available vm names\n")
	fmt.Printf("       moderniedownlaoder -h     print this help\n")
}
