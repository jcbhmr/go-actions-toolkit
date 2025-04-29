package platform

import (
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

func getWindowsInfo() (struct {
	Name    string
	Version string
}, error) {
	cmd := exec.Command("powershell", "-command", "(Get-CimInstance -ClassName Win32_OperatingSystem).Version")
	version, err := cmd.Output()
	if err != nil {
		return struct {
			Name    string
			Version string
		}{}, err
	}

	cmd = exec.Command("powershell", "-command", "(Get-CimInstance -ClassName Win32_OperatingSystem).Caption")
	name, err := cmd.Output()
	if err != nil {
		return struct {
			Name    string
			Version string
		}{}, err
	}

	return struct {
		Name    string
		Version string
	}{
		Version: strings.TrimSpace(string(version)),
		Name:    strings.TrimSpace(string(name)),
	}, nil
}

func getMacOSInfo() (struct {
	Name    string
	Version string
}, error) {
	cmd := exec.Command("sw_vers")
	out, err := cmd.Output()
	if err != nil {
		return struct {
			Name    string
			Version string
		}{}, err
	}
	version := func() string {
		versionMatch := regexp.MustCompile(`ProductVersion:\s*(.+)`).FindStringSubmatch(string(out))
		if len(versionMatch) == 0 {
			return ""
		}
		return versionMatch[1]
	}()
	name := func() string {
		nameMatch := regexp.MustCompile(`ProductName:\s*(.+)`).FindStringSubmatch(string(out))
		if len(nameMatch) == 0 {
			return ""
		}
		return nameMatch[1]
	}()

	return struct {
		Name    string
		Version string
	}{
		Name:    name,
		Version: version,
	}, nil
}

func getLinuxInfo() (struct {
	Name    string
	Version string
}, error) {
	cmd := exec.Command("lsb_release", "-i", "-r", "-s")
	out, err := cmd.Output()
	if err != nil {
		return struct {
			Name    string
			Version string
		}{}, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return struct {
			Name    string
			Version string
		}{}, nil
	}
	name := strings.TrimSpace(lines[0])
	version := strings.TrimSpace(lines[1])

	return struct {
		Name    string
		Version string
	}{
		Name:    name,
		Version: version,
	}, nil
}

const Platform = runtime.GOOS
const Arch = runtime.GOARCH
const IsWindows = Platform == "windows"
const IsMacOS = Platform == "darwin"
const IsLinux = Platform == "linux"

func GetDetails() (struct {
	Name      string
	Platform  string
	Arch      string
	Version   string
	IsWindows bool
	IsMacOS   bool
	IsLinux   bool
}, error) {
	var err error
	var info struct {
		Name    string
		Version string
	}
	if IsWindows {
		info, err = getWindowsInfo()
		if err != nil {
			return struct {
				Name      string
				Platform  string
				Arch      string
				Version   string
				IsWindows bool
				IsMacOS   bool
				IsLinux   bool
			}{}, err
		}
	} else if IsMacOS {
		info, err = getMacOSInfo()
		if err != nil {
			return struct {
				Name      string
				Platform  string
				Arch      string
				Version   string
				IsWindows bool
				IsMacOS   bool
				IsLinux   bool
			}{}, err
		}
	} else {
		info, err = getLinuxInfo()
		if err != nil {
			return struct {
				Name      string
				Platform  string
				Arch      string
				Version   string
				IsWindows bool
				IsMacOS   bool
				IsLinux   bool
			}{}, err
		}
	}
	return struct {
		Name      string
		Platform  string
		Arch      string
		Version   string
		IsWindows bool
		IsMacOS   bool
		IsLinux   bool
	}{
		Name:      info.Name,
		Platform:  Platform,
		Arch:      Arch,
		Version:   info.Version,
		IsWindows: IsWindows,
		IsMacOS:   IsMacOS,
		IsLinux:   IsLinux,
	}, nil
}
