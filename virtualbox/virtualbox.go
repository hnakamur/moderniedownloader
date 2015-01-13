package virtualbox

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/hnakamur/moderniedownloader/executil"
	"github.com/hnakamur/moderniedownloader/vmlist"
)

const (
	BrowserOSSeparator         = " - "
	firstSnapShotName          = "Snapshot 1"
	ClipboardModeBidirectional = "bidirectional"
)

func DoesVmExist(vmName string) (bool, error) {
	cmd := exec.Command("VBoxManage", "showvminfo", vmName)
	exitStatus, err := executil.Run(cmd)
	if err != nil {
		return false, err
	}
	if exitStatus.ExitCode == 0 {
		return true, nil
	} else {
		return false, nil
	}
}

var osVersionMappingFromVmNameToVmList = map[string]string{
	"WinXP":  "XP",
	"Vista":  "Vista",
	"Win7":   "Win7",
	"Win8":   "Win8",
	"Win8.1": "Win8.1",
	"Win10":  "Win10",
}

var osVersionMappingFromVmListToVmName map[string]string

func init() {
	osVersionMappingFromVmListToVmName = make(map[string]string)
	for k, v := range osVersionMappingFromVmNameToVmList {
		osVersionMappingFromVmListToVmName[v] = k
	}
}

func GetRegisteredVmNameList() ([]string, error) {
	cmd := exec.Command("VBoxManage", "list", "vms")
	var out bytes.Buffer
	cmd.Stdout = &out

	exitStatus, err := executil.Run(cmd)
	if err != nil {
		return nil, err
	}
	if exitStatus.ExitCode != 0 {
		return nil, fmt.Errorf("VBoxManage list vms failed with exitCode=%d", exitStatus.ExitCode)
	}

	var vmNames []string
	re := regexp.MustCompile(`^"(IE[\d.]+ - Win[\d.]+)"`)
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		line := scanner.Text()
		values := re.FindAllStringSubmatch(line, 1)
		if values != nil {
			vmName := values[0][1]
			vmNames = append(vmNames, vmName)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("Error during reading output of VBoxManage list vms: %s", err)
	}

	return vmNames, nil
}

func GetVmNameList() ([]string, error) {
	browsers, err := vmlist.GetBrowsers("Mac", "VirtualBox")
	if err != nil {
		return nil, err
	}

	vmNames := make([]string, len(browsers))
	for i, browser := range browsers {
		vmNames[i] = fmt.Sprintf("IE%s - %s", browser.Version, osVersionMappingFromVmListToVmName[browser.OsVersion])
	}
	return vmNames, nil
}

func NewVmListBrowserSpecFromVmName(vmName string) (*vmlist.BrowserSpec, error) {
	browserVersion := getBrowserVersionFromVMName(vmName)
	if browserVersion == "" {
		return nil, fmt.Errorf("Invalid browserVersion in vmName: %s", vmName)
	}

	osVersionInVMName := getOSVersionFromVMName(vmName)
	osVersion := osVersionMappingFromVmNameToVmList[osVersionInVMName]
	if osVersion == "" {
		return nil, fmt.Errorf("Unknown osVersion in vmName: %s", vmName)
	}

	return &vmlist.BrowserSpec{
		OsName:       "Mac",
		SoftwareName: "VirtualBox",
		Version:      browserVersion,
		OsVersion:    osVersion,
	}, nil
}

func StartVm(vmName string) error {
	cmd := exec.Command("VBoxManage", "startvm", vmName, "--type", "gui")
	exitStatus, err := executil.Run(cmd)
	if err != nil {
		return err
	}
	if exitStatus.ExitCode == 0 {
		return nil
	} else {
		return fmt.Errorf("VBoxManage startvm \"%s\" failed with exitCode=%d", vmName, exitStatus.ExitCode)
	}
}

func ImportAndConfigureVm(vmName string) error {
	err := importVm(GetOvaFileNameForVmName(vmName))
	if err != nil {
		return err
	}

	err = configVmMemory(vmName)
	if err != nil {
		return err
	}

	err = attachGuestAdditionsMedia(vmName)
	if err != nil {
		return err
	}

	err = takeSnapshot(vmName, firstSnapShotName)
	if err != nil {
		return err
	}

	return nil
}

func getBrowserVersionFromVMName(vmName string) string {
	i := strings.Index(vmName, BrowserOSSeparator)
	if i == -1 {
		return ""
	}

	prefix := "IE"
	prefixLen := len(prefix)
	if vmName[:prefixLen] != prefix {
		return ""
	}
	return vmName[prefixLen:i]
}

func getOSVersionFromVMName(vmName string) string {
	i := strings.Index(vmName, BrowserOSSeparator)
	if i == -1 {
		return ""
	}
	return vmName[i+len(BrowserOSSeparator):]
}

func GetOvaFileNameForVmName(vmName string) string {
	return vmName + ".ova"
}

func importVm(ovaFilename string) error {
	cmd := exec.Command("VBoxManage", "import", ovaFilename)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	exitStatus, err := executil.Run(cmd)
	if err != nil {
		return err
	}
	if exitStatus.ExitCode == 0 {
		return nil
	} else {
		return fmt.Errorf("VBoxManage import \"%s\" failed with exitCode=%d", ovaFilename, exitStatus.ExitCode)
	}
}

func configVmMemory(vmName string) error {
	osVersion := getOSVersionFromVMName(vmName)
	var cmd *exec.Cmd
	if osVersion == "WinXP" || osVersion == "Vista" {
		cmd = exec.Command("VBoxManage", "modifyvm", vmName, "--memory", "1024")
	} else if osVersion == "Win7" || osVersion == "Win8" || osVersion == "Win8.1" {
		cmd = exec.Command("VBoxManage", "modifyvm", vmName, "--memory", "2048", "--vram", "128")
	} else {
		return fmt.Errorf("Unsupported os version: %s", osVersion)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	exitStatus, err := executil.Run(cmd)
	if err != nil {
		return err
	}
	if exitStatus.ExitCode == 0 {
		return nil
	} else {
		return fmt.Errorf("VBoxManage import \"%s\" failed with exitCode=%d", vmName, exitStatus.ExitCode)
	}
}

func attachGuestAdditionsMedia(vmName string) error {
	cmd := exec.Command("VBoxManage", "storageattach", vmName, "--storagectl", "IDE", "--port", "1", "--device", "0", "--type", "dvddrive", "--medium", "additions")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	exitStatus, err := executil.Run(cmd)
	if err != nil {
		return err
	}
	if exitStatus.ExitCode == 0 {
		return nil
	} else {
		return fmt.Errorf("VBoxManage storageattach for VM \"%s\" failed with exitCode=%d", vmName, exitStatus.ExitCode)
	}
}

func SetClipboardMode(vmName, clipboardMode string) error {
	cmd := exec.Command("VBoxManage", "controlvm", vmName, "clipboard", clipboardMode)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	exitStatus, err := executil.Run(cmd)
	if err != nil {
		return err
	}
	if exitStatus.ExitCode == 0 {
		return nil
	} else {
		return fmt.Errorf("VBoxManage controlvm for setting clipboard mode of VM \"%s\" failed with exitCode=%d", vmName, exitStatus.ExitCode)
	}
}

func takeSnapshot(vmName, snapshotName string) error {
	cmd := exec.Command("VBoxManage", "snapshot", vmName, "take", snapshotName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	exitStatus, err := executil.Run(cmd)
	if err != nil {
		return err
	}
	if exitStatus.ExitCode == 0 {
		return nil
	} else {
		return fmt.Errorf("VBoxManage snapshot for VM \"%s\" failed with exitCode=%d", vmName, exitStatus.ExitCode)
	}
}
