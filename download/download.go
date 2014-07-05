package download

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"sync"

	"github.com/hnakamur/moderniedownloader/executil"
	"github.com/hnakamur/moderniedownloader/virtualbox"
	"github.com/hnakamur/moderniedownloader/vmlist"
)

func DoesOvaFileExist(vmName string) (bool, error) {
	return fileExists(virtualbox.GetOvaFileNameForVmName(vmName))
}

func DownloadAndBuildOvaFile(files []vmlist.ChunkFile) error {
	downloadFilesIfNeeded(files)

	err := concatFiles(files)
	if err != nil {
		return err
	}

	return removeFiles(files)
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

func downloadFilesIfNeeded(files []vmlist.ChunkFile) {
	var wg sync.WaitGroup
	wg.Add(len(files))
	for i, file := range files {
		go func(fileId int, f vmlist.ChunkFile) {
			downloadMd5AndFileIfMd5NotMatch(f.Md5url, f.Url, f.GetLocalFileName())
			wg.Done()
		}(i, file)
	}
	wg.Wait()
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

func concatFiles(files []vmlist.ChunkFile) error {
	executableFileName := files[0].GetLocalFileName()
	fmt.Printf("chmod +x %s", executableFileName)
	cmd := exec.Command("chmod", "+x", executableFileName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	exitStatus, err := executil.Run(cmd)
	if err != nil {
		return err
	}
	if exitStatus.ExitCode != 0 {
		return fmt.Errorf("chmod +x %s failed with exitCode=%d", executableFileName, exitStatus.ExitCode)
	}

	fmt.Printf("Running ./%s", executableFileName)
	cmd = exec.Command("./" + executableFileName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	exitStatus, err = executil.Run(cmd)
	if err != nil {
		return err
	}
	if exitStatus.ExitCode != 0 {
		return fmt.Errorf("./% failed with exitCode=%d", executableFileName, exitStatus.ExitCode)
	}

	return nil
}

func removeFiles(files []vmlist.ChunkFile) error {
	for _, file := range files {
		err := os.Remove(file.GetLocalFileName())
		if err != nil {
			return err
		}
	}
	return nil
}
