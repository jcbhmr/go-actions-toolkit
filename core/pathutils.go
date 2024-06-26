package core

import (
	"runtime"
	"strings"
)

// toPosixPath converts the given path to the posix form. On Windows, \\ will be
// replaced with /.
//
// @param pth. Path to transform.
// @return string Posix path.
func ToPosixPath(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

// toWin32Path converts the given path to the win32 form. On Linux, / will be
// replaced with \\.
//
// @param pth. Path to transform.
// @return string Win32 path.
func ToWin32Path(path string) string {
	return strings.ReplaceAll(path, "/", "\\")
}

// toPlatformPath converts the given path to a platform-specific path. It does
// this by replacing instances of / and \ with the platform-specific path
// separator.
//
// @param pth The path to platformize.
// @return string The platform-specific path.
func ToPlatformPath(path string) string {
	if runtime.GOOS == "windows" {
		return ToWin32Path(path)
	} else {
		return ToPosixPath(path)
	}
}
