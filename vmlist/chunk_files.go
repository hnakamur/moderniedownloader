package vmlist

import (
	"encoding/json"
	"fmt"
	"io"
	"path"
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
	Browsers     []browserData
}

type browserData struct {
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

func GetFilesForBrowser(r io.Reader, spec *BrowserSpec) ([]ChunkFile, error) {
	var osList []osData
	decoder := json.NewDecoder(r)
	err := decoder.Decode(&osList)
	if err != nil {
		return nil, err
	}

	softwareList := getSoftwareListForOsName(osList, spec.OsName)
	if softwareList == nil {
		return nil, fmt.Errorf("softwareList not found for os: %s", spec.OsName)
	}

	browsers := getBrowsersForSoftwareName(softwareList, spec.SoftwareName)
	if browsers == nil {
		return nil, fmt.Errorf("browsers not found for softwareName: %s", spec.SoftwareName)
	}

	files := getFilesForVersionAndOsVersion(browsers, spec.Version, spec.OsVersion)
	if files == nil {
		return nil, fmt.Errorf("files not found for version: %s, osVersion: %s", spec.Version, spec.OsVersion)
	}

	return files, nil
}

func getSoftwareListForOsName(osList []osData, osName string) []softwareData {
	for _, os := range osList {
		if os.OsName == osName {
			return os.SoftwareList
		}
	}
	return nil
}

func getBrowsersForSoftwareName(softwareList []softwareData, softwareName string) []browserData {
	for _, software := range softwareList {
		if software.SoftwareName == softwareName {
			return software.Browsers
		}
	}
	return nil
}

func getFilesForVersionAndOsVersion(browsers []browserData, version, osVersion string) []ChunkFile {
	for _, browser := range browsers {
		if browser.Version == version && browser.OsVersion == osVersion {
			return browser.Files
		}
	}
	return nil
}
