package download

import (
	"archive/zip"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/hnakamur/moderniedownloader/virtualbox"
	"github.com/hnakamur/moderniedownloader/vmlist"
)

func DoesOvaFileExist(vmName string) (bool, error) {
	return fileExists(virtualbox.GetOvaFileNameForVmName(vmName))
}

func DownloadAndBuildOvaFile(file vmlist.ChunkFile) error {
	if err := downloadFileIfNeeded(file); err != nil {
		return err
	}

	filename := file.GetLocalFileName()
	if err := unzipFile(filename); err != nil {
		return err
	}

	return os.Remove(filename)
}

func fileExists(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	defer f.Close()
	return true, nil
}

func downloadFileIfNeeded(f vmlist.ChunkFile) error {
	return downloadMd5AndFileIfMd5NotMatch(f.Md5url, f.Url, f.GetLocalFileName())
}

func downloadMd5AndFileIfMd5NotMatch(md5url, url, path string) error {
	md5, err := fetchMd5(md5url)
	if err != nil {
		return err
	}

	return downloadFileIfMd5NotMatch(md5, url, path)
}

func fetchMd5(md5url string) (string, error) {
	resp, err := http.Get(md5url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func downloadFileIfMd5NotMatch(md5, url, path string) error {
	localMd5, err := calcMd5OfFile(path)
	if err != nil {
		return err
	}
	if localMd5 == md5 {
		return nil
	}

	fmt.Printf("Start downloading %s from url %s ...\n", path, url)
	localMd5, err = downloadFileAndCalcMd5(url, path)
	if err != nil {
		return err
	}
	if localMd5 != md5 {
		return fmt.Errorf("Md5 unmatched. remote=%s, local=%s",
			md5, localMd5)
	}
	fmt.Printf("Finished downloading %s from url %s\n", path, url)

	return nil
}

func downloadFileAndCalcMd5(url, path string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	writer, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer writer.Close()

	h := md5.New()
	reader := io.TeeReader(resp.Body, h)
	_, err = io.Copy(writer, reader)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func calcMd5OfFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	defer f.Close()
	return calcMd5(f)
}

func calcMd5(rd io.Reader) (string, error) {
	h := md5.New()
	_, err := io.Copy(h, rd)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func unzipFile(filename string) error {
	r, err := zip.OpenReader(filename)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Mode()&os.ModeDir != 0 {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		out, err := os.Create(path.Base(f.Name))
		if err != nil {
			return nil
		}
		defer out.Close()

		if _, err = io.Copy(out, rc); err != nil {
			return err
		}
	}
	return nil
}
