package core

import (
	"os/exec"
	"runtime"
	"strings"
)

var PlatformPlatform =  runtime.GOOS
var PlatformArch =      runtime.GOARCH
var PlatformIsWindows = runtime.GOOS == "windows"
var PlatformIsMacOs =   runtime.GOOS == "darwin"
var PlatformIsLinux =   runtime.GOOS == "linux"

type platformDetails struct {
	Name      string
	Platform  string
	Arch      string
	Version   string
	IsWindows bool
	IsMacOs   bool
	IsLinux   bool
}

func platformGetWindowsInfo() (string, string, error) {
	version, err := exec.Command("powershell", "-command", "(Get-CimInstance -ClassName Win32_OperatingSystem).Version").Output()
	if err != nil {
		return "", "", err
	}
	versionStr := strings.TrimSpace(string(version))
	name, err := exec.Command("powershell", "-command", "(Get-CimInstance -ClassName Win32_OperatingSystem).Caption").Output()
	if err != nil {
		return "", "", err
	}
	nameStr := strings.TrimSpace(string(name))
	return nameStr, versionStr, nil
}

func platformGetMacOsInfo() (string, string, error) {
	version, err := exec.Command("sw_vers", "-productVersion").Output()
	if err != nil {
		return "", "", err
	}
	versionStr := strings.TrimSpace(string(version))
	name, err := exec.Command("sw_vers", "-productName").Output()
	if err != nil {
		return "", "", err
	}
	nameStr := strings.TrimSpace(string(name))
	return nameStr, versionStr, nil
}

func platformGetLinuxInfo() (string, string, error) {
	name, err := exec.Command("lsb_release", "-is").Output()
	if err != nil {
		return "", "", err
	}
	nameStr := strings.TrimSpace(string(name))
	version, err := exec.Command("lsb_release", "-rs").Output()
	if err != nil {
		return "", "", err
	}
	versionStr := strings.TrimSpace(string(version))
	return nameStr, versionStr, nil
}

func Platform_GetDetails() (*platformDetails, error) {
	var name string
	var version string
	var err error
	if PlatformIsWindows {
		name, version, err = platformGetWindowsInfo()
	} else if PlatformIsMacOs {
		name, version, err = platformGetMacOsInfo()
	} else if PlatformIsLinux {
		name, version, err = platformGetLinuxInfo()
	}
	if err != nil {
		return nil, err
	}
	return &platformDetails{
		Name:      name,
		Version:   version,
		Platform:  PlatformPlatform,
		Arch:      PlatformArch,
		IsWindows: PlatformIsWindows,
		IsMacOs:   PlatformIsMacOs,
		IsLinux:   PlatformIsLinux,
	}, nil
}
