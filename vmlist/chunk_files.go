package vmlist

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/hnakamur/moderniedownloader/scraping"
)

type BrowserSpec struct {
	OsName       string
	SoftwareName string
	Version      string
	OsVersion    string
}

type osData struct {
	OsName       string
	SoftwareList []softwareData
}

type softwareData struct {
	SoftwareName string
	Browsers     []BrowserData
}

type BrowserData struct {
	Version   string
	OsVersion string
	Files     []ChunkFile
}

type ChunkFile struct {
	Md5url string `json:"md5"`
	Url    string
}

func (f *ChunkFile) GetLocalFileName() string {
	return path.Base(f.Url)
}

func GetFilesForBrowser(spec *BrowserSpec) ([]ChunkFile, error) {
	browsers, err := GetBrowsers(spec.OsName, spec.SoftwareName)
	if err != nil {
		return nil, err
	}

	files := getFilesForVersionAndOsVersion(browsers, spec.Version, spec.OsVersion)
	if files == nil {
		return nil, fmt.Errorf("files not found for version: %s, osVersion: %s", spec.Version, spec.OsVersion)
	}

	return files, nil
}

func GetBrowsers(osName, softwareName string) ([]BrowserData, error) {
	osList, err := downloadOsList()
	if err != nil {
		return nil, err
	}

	softwareList := getSoftwareListForOsName(osList, osName)
	if softwareList == nil {
		return nil, fmt.Errorf("softwareList not found for os: %s", osName)
	}

	browsers := getBrowsersForSoftwareName(softwareList, softwareName)
	if browsers == nil {
		return nil, fmt.Errorf("browsers not found for softwareName: %s", softwareName)
	}

	return browsers, nil
}

func downloadOsList() ([]osData, error) {
	list, err := scraping.DownloadVmOsList()
	if err != nil {
		return nil, err
	}

	var osList []osData
	decoder := json.NewDecoder(strings.NewReader(list))
	err = decoder.Decode(&osList)
	if err != nil {
		return nil, err
	}

	return osList, nil
}

func getSoftwareListForOsName(osList []osData, osName string) []softwareData {
	for _, os := range osList {
		if os.OsName == osName {
			return os.SoftwareList
		}
	}
	return nil
}

func getBrowsersForSoftwareName(softwareList []softwareData, softwareName string) []BrowserData {
	for _, software := range softwareList {
		if software.SoftwareName == softwareName {
			return software.Browsers
		}
	}
	return nil
}

func getFilesForVersionAndOsVersion(browsers []BrowserData, version, osVersion string) []ChunkFile {
	for _, browser := range browsers {
		if browser.Version == version && browser.OsVersion == osVersion {
			return browser.Files
		}
	}
	return nil
}
