package download

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func DownloadMd5AndFileIfMd5NotMatch(md5url, url, path string) error {
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

	localMd5, err = downloadFileAndCalcMd5(url, path)
	if localMd5 != md5 {
		return fmt.Errorf("Md5 unmatched. remote=%s, local=%s",
			md5, localMd5)
	}

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
