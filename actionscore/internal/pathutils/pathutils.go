package pathutils

import (
	"path/filepath"
	"strings"
)

func ToPosixPath(p string) string {
	return strings.ReplaceAll(p, "\\", "/")
}

func ToWin32Path(p string) string {
	return strings.ReplaceAll(p, "/", "\\")
}

func ToPlatformPath(p string) string {
	return strings.NewReplacer("/", string(filepath.Separator), "\\", string(filepath.Separator)).Replace(p)
}
