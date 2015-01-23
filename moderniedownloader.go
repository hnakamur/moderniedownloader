package main

import (
	"flag"
	"fmt"

	"github.com/hnakamur/moderniedownloader/download"
	"github.com/hnakamur/moderniedownloader/virtualbox"
	"github.com/hnakamur/moderniedownloader/vmlist"
)

var listRegisteredVmsFlag bool
var listAvailableVmsFlag bool
var helpFlag bool

func init() {
	flag.BoolVar(&listRegisteredVmsFlag, "l", false, "list registered modern.IE VM names")
	flag.BoolVar(&listAvailableVmsFlag, "L", false, "list available modern.IE VM names")
	flag.BoolVar(&helpFlag, "h", false, "help")
}

func main() {
	flag.Parse()

	if listRegisteredVmsFlag {
		listRegisteredVmNames()
		return
	}
	if listAvailableVmsFlag {
		listAvailableVmNames()
		return
	}
	if helpFlag || flag.NArg() == 0 {
		usage()
		return
	}

	vmName := flag.Arg(0)

	vmExists, err := virtualbox.DoesVmExist(vmName)
	if err != nil {
		panic(err)
	}

	if !vmExists {
		err = setupVM(vmName)
		if err != nil {
			panic(err)
		}
	}

	err = virtualbox.StartVm(vmName)
	if err != nil {
		panic(err)
	}

	if !vmExists {
		err = virtualbox.SetClipboardMode(vmName, virtualbox.ClipboardModeBidirectional)
		if err != nil {
			panic(err)
		}
	}
}

func setupVM(vmName string) error {
	ovaFileExists, err := download.DoesOvaFileExist(vmName)
	if err != nil {
		return err
	}

	if !ovaFileExists {
		err = downloadAndBuildOvaFile(vmName)
		if err != nil {
			return err
		}
	}

	return virtualbox.ImportAndConfigureVm(vmName)
}

func downloadAndBuildOvaFile(vmName string) error {
	spec, err := virtualbox.NewVmListBrowserSpecFromVmName(vmName)
	if err != nil {
		return err
	}

	f, err := vmlist.GetFileInfoForBrowser(spec)
	if err != nil {
		return err
	}

	return download.DownloadAndBuildOvaFile(f)
}

func listRegisteredVmNames() {
	vmNames, err := virtualbox.GetRegisteredVmNameList()
	if err != nil {
		panic(err)
	}
	for _, vmName := range vmNames {
		fmt.Printf("%s\n", vmName)
	}
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
	fmt.Printf("       moderniedownlaoder -l     list registered vm names\n")
	fmt.Printf("       moderniedownlaoder -L     list available vm names\n")
	fmt.Printf("       moderniedownlaoder -h     print this help\n")
}
