package vmwareworkstation

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"github.com/docker/machine/libmachine/log"
)

func detectVMRunCmd() string {
	if path := findFile("vmrun.exe", workstationProgramFilePaths()); path != "" {
		return path
	}

	return detectCmdInPath("vmrun.exe")
}

func detectVDiskManCmd() string {
	if path := findFile("vmware-vdiskmanager.exe", workstationProgramFilePaths()); path != "" {
		return path
	}

	return detectCmdInPath("vmware-vdiskmanager.exe")
}

func detectDhcpLeasesPath() string {
	path, err := workstationDhcpLeasesPathRegistry()
	if err != nil {
		log.Errorf("Error finding leases in registry: %s", err)
	} else if _, err := os.Stat(path); err == nil {
		return path
	}

	return findFile("vmnetdhcp.leases", workstationDataFilePaths())
}

func normalizePath(path string) string {
	path = strings.Replace(path, "\\", "/", -1)
	path = strings.Replace(path, "//", "/", -1)
	path = strings.TrimRight(path, "/")
	return path
}

func findFile(file string, paths []string) string {
	for _, path := range paths {
		path = filepath.Join(path, file)
		path = normalizePath(path)
		log.Debugf("Searching for file '%s'", path)

		if _, err := os.Stat(path); err == nil {
			log.Debugf("Found file '%s'", path)
			return path
		}
	}

	log.Errorf("File not found: '%s'", file)
	return ""
}

// See http://blog.natefinch.com/2012/11/go-win-stuff.html
//
func readRegString(hive syscall.Handle, subKeyPath, valueName string) (value string, err error) {
	var h syscall.Handle
	err = syscall.RegOpenKeyEx(hive, syscall.StringToUTF16Ptr(subKeyPath), 0, syscall.KEY_READ, &h)
	if err != nil {
		return
	}
	defer syscall.RegCloseKey(h)

	var typ uint32
	var bufSize uint32
	err = syscall.RegQueryValueEx(
		h,
		syscall.StringToUTF16Ptr(valueName),
		nil,
		&typ,
		nil,
		&bufSize)
	if err != nil {
		return
	}

	data := make([]uint16, bufSize/2+1)
	err = syscall.RegQueryValueEx(
		h,
		syscall.StringToUTF16Ptr(valueName),
		nil,
		&typ,
		(*byte)(unsafe.Pointer(&data[0])),
		&bufSize)
	if err != nil {
		return
	}

	return syscall.UTF16ToString(data), nil
}

// This reads the VMware DHCP leases path from the Windows registry.
func workstationDhcpLeasesPathRegistry() (s string, err error) {
	key := `SYSTEM\CurrentControlSet\services\VMnetDHCP\Parameters`
	subkey := "LeaseFile"
	s, err = readRegString(syscall.HKEY_LOCAL_MACHINE, key, subkey)
	if err != nil {
		log.Errorf(`Unable to read registry key %s\%s`, key, subkey)
		return
	}

	return normalizePath(s), nil
}

// workstationDataFilePaths returns a list of paths that are eligible
// to contain data files we may want such as vmnet NAT configuration files.
func workstationDataFilePaths() []string {
	leasesPath, err := workstationDhcpLeasesPathRegistry()
	if err != nil {
		log.Errorf("Error getting DHCP leases path: %s", err)
	}

	if leasesPath != "" {
		leasesPath = filepath.Dir(leasesPath)
	}

	paths := make([]string, 0, 5)
	if os.Getenv("VMWARE_DATA") != "" {
		paths = append(paths, os.Getenv("VMWARE_DATA"))
	}

	if leasesPath != "" {
		paths = append(paths, leasesPath)
	}

	if os.Getenv("ProgramData") != "" {
		paths = append(paths,
			filepath.Join(os.Getenv("ProgramData"), "/VMware"))
	}

	if os.Getenv("ALLUSERSPROFILE") != "" {
		paths = append(paths,
			filepath.Join(os.Getenv("ALLUSERSPROFILE"), "/Application Data/VMware"))
	}

	return paths
}

// This reads the VMware installation path from the Windows registry.
func workstationVMwareRoot() (s string, err error) {
	key := `SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\vmware.exe`
	subkey := "Path"
	s, err = readRegString(syscall.HKEY_LOCAL_MACHINE, key, subkey)
	if err != nil {
		log.Errorf(`Unable to read registry key %s\%s`, key, subkey)
		return
	}

	return normalizePath(s), nil
}

// workstationProgramFilesPaths returns a list of paths that are eligible
// to contain program files we may want just as vmware.exe.
func workstationProgramFilePaths() []string {
	path, err := workstationVMwareRoot()
	if err != nil {
		log.Errorf("Error finding VMware root: %s", err)
	}

	paths := make([]string, 0, 5)
	if os.Getenv("VMWARE_HOME") != "" {
		paths = append(paths, os.Getenv("VMWARE_HOME"))
	}

	if path != "" {
		paths = append(paths, path)
	}

	if os.Getenv("ProgramFiles(x86)") != "" {
		paths = append(paths,
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "/VMware/VMware Workstation"))
	}

	if os.Getenv("ProgramFiles") != "" {
		paths = append(paths,
			filepath.Join(os.Getenv("ProgramFiles"), "/VMware/VMware Workstation"))
	}

	return paths
}
