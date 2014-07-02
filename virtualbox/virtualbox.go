package virtualbox

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/hnakamur/moderniedownloader/executil"
	"github.com/hnakamur/moderniedownloader/vmlist"
)

const (
	BrowserOSSeparator = " - "
	FirstSnapShotName  = "Snapshot 1"
	ClipboardMode      = "bidirectional"
)

func DoesVMExist(vmName string) (bool, error) {
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

var osVersionMappingFromVMNameToVMList = map[string]string{
	"WinXP":  "XP",
	"Vista":  "vista",
	"Win7":   "win7",
	"Win8":   "win8",
	"Win8.1": "win8.1",
}

var osVersionMappingFromVMListToVMName map[string]string

func init() {
	osVersionMappingFromVMListToVMName = make(map[string]string)
	for k, v := range osVersionMappingFromVMNameToVMList {
		osVersionMappingFromVMListToVMName[v] = k
	}
}

func GetVmNameList() ([]string, error) {
	browsers, err := vmlist.GetBrowsers("mac", "virtualbox")
	if err != nil {
		return nil, err
	}

	vmNames := make([]string, len(browsers))
	for i, browser := range browsers {
		vmNames[i] = fmt.Sprintf("IE%s - %s", browser.Version, osVersionMappingFromVMListToVMName[browser.OsVersion])
	}
	return vmNames, nil
}

func NewVMListBrowserSpecFromVMName(vmName string) (*vmlist.BrowserSpec, error) {
	browserVersion := getBrowserVersionFromVMName(vmName)
	if browserVersion == "" {
		return nil, fmt.Errorf("Invalid browserVersion in vmName: %s", vmName)
	}

	osVersionInVMName := getOSVersionFromVMName(vmName)
	osVersion := osVersionMappingFromVMNameToVMList[osVersionInVMName]
	if osVersion == "" {
		return nil, fmt.Errorf("Unknown osVersion in vmName: %s", vmName)
	}

	return &vmlist.BrowserSpec{
		OsName:       "mac",
		SoftwareName: "virtualbox",
		Version:      browserVersion,
		OsVersion:    osVersion,
	}, nil
}

func StartVM(vmName string) error {
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

func ImportAndConfigureVM(vmName string) error {
	err := importVM(GetOVAFilenameForVMName(vmName))
	if err != nil {
		return err
	}

	err = configVMMemory(vmName)
	if err != nil {
		return err
	}

	err = attachGuestAdditionsMedia(vmName)
	if err != nil {
		return err
	}

	err = setClipboardMode(vmName, ClipboardMode)
	if err != nil {
		return err
	}

	err = takeSnapshot(vmName, FirstSnapShotName)
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

func GetOVAFilenameForVMName(vmName string) string {
	return vmName + ".ova"
}

func importVM(ovaFilename string) error {
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

func configVMMemory(vmName string) error {
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

func setClipboardMode(vmName, clipboardMode string) error {
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
