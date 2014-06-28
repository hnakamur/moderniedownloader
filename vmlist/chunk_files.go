package vmlist

import (
	"fmt"
	"io"

	simplejson "github.com/bitly/go-simplejson"
)

type BrowserSpec struct {
	OsName       string
	SoftwareName string
	Version      string
	OsVersion    string
}

type ChunkFile struct {
	Md5url string
	Url    string
}

func GetFilesForBrowser(r io.Reader, spec *BrowserSpec) ([]ChunkFile, error) {
	osList, err := simplejson.NewFromReader(r)
	if err != nil {
		return nil, err
	}

	softwareList, err := getSoftwareListForOsName(osList, spec.OsName)
	if err != nil {
		return nil, err
	}

	browsers, err := getBrowsersForSoftwareName(softwareList, spec.SoftwareName)
	if err != nil {
		return nil, err
	}

	return getFilesForVersionAndOsVersion(browsers, spec.Version, spec.OsVersion)
}

func getSoftwareListForOsName(js *simplejson.Json, osName string) (*simplejson.Json, error) {
	return findArrayForKey(js, "softwareList", "osName", osName)
}

func getBrowsersForSoftwareName(js *simplejson.Json, softwareName string) (*simplejson.Json, error) {
	return findArrayForKey(js, "browsers", "softwareName", softwareName)
}

func getFilesForVersionAndOsVersion(js *simplejson.Json, version, osVersion string) ([]ChunkFile, error) {
	jsFiles, err := findArrayFor2Keys(js, "files", "version", version, "osVersion", osVersion)
	if err != nil {
		return nil, err
	}

	elems, err := jsFiles.Array()
	if err != nil {
		return nil, err
	}

	files := make([]ChunkFile, len(elems))
	for i, _ := range elems {
		elem := jsFiles.GetIndex(i)
		if err != nil {
			return nil, err
		}

		md5url, err := elem.Get("md5").String()
		if err != nil {
			return nil, err
		}

		url, err := elem.Get("url").String()
		if err != nil {
			return nil, err
		}

		files[i] = ChunkFile{
			Md5url: md5url,
			Url:    url,
		}
	}
	return files, nil
}

func findArrayForKey(js *simplejson.Json, arrayKey, keyName, keyValue string) (*simplejson.Json, error) {
	elems, err := js.Array()
	if err != nil {
		return nil, err
	}

	for i, _ := range elems {
		elem := js.GetIndex(i)
		if err != nil {
			return nil, err
		}

		value, err := elem.Get(keyName).String()
		if err != nil {
			return nil, err
		}

		if value == keyValue {
			jsArray := elem.Get(arrayKey)
			_, err := jsArray.Array()
			if err != nil {
				return nil, err
			}

			return jsArray, nil
		}
	}
	return nil, fmt.Errorf("%s:\"%s\" not found", keyName, keyValue)
}

func findArrayFor2Keys(js *simplejson.Json, arrayKey, key1Name, key1Value, key2Name, key2Value string) (*simplejson.Json, error) {
	elems, err := js.Array()
	if err != nil {
		return nil, err
	}

	for i, _ := range elems {
		elem := js.GetIndex(i)
		if err != nil {
			return nil, err
		}

		value1, err := elem.Get(key1Name).String()
		if err != nil {
			return nil, err
		}

		value2, err := elem.Get(key2Name).String()
		if err != nil {
			return nil, err
		}

		if value1 == key1Value && value2 == key2Value {
			jsArray := elem.Get(arrayKey)
			_, err := jsArray.Array()
			if err != nil {
				return nil, err
			}

			return jsArray, nil
		}
	}
	return nil, fmt.Errorf("%s:\"%s\", %s:\"%s\" not found", key1Name, key1Value, key2Name, key2Value)
}
