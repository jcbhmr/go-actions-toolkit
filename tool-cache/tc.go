package tc

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/google/uuid"
)

type HttpError interface {
	error
}

func NewHttpError(httpStatusCode *int) HttpError {
	return fmt.Errorf("unexpected HTTP response: %v", httpStatusCode)
}

func ptr[T any](v T) *T {
	return &v
}

func DownloadTool(url string, dest *string, auth *string, headers *map[string]string) (string, error) {
	var dest2 string
	if dest == nil {
		dest2 = filepath.Join(os.Getenv("RUNNER_TEMP"), uuid.NewString())
	} else {
		dest2 = *dest
	}
	err := os.MkdirAll(filepath.Dir(dest2), 0o755)
	if err != nil {
		return "", err
	}
	fmt.Printf("::debug::downloading %s", url)
	fmt.Printf("::debug::destination %s", dest2)
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	if res.StatusCode != 200 {
		return "", NewHttpError(&res.StatusCode)
	}
	defer res.Body.Close()
	out, err := os.Create(dest2)
	if err != nil {
		return "", err
	}
	defer out.Close()
	_, err = io.Copy(out, res.Body)
	if err != nil {
		return "", err
	}
	return dest2, nil
}

func Extract7z(file string, dest *string, the7zPath *string) (string, error) {
	if runtime.GOOS != "windows" {
		return "", errors.New("Extract7z not supported on current OS")
	}
	dest2, err := createExtractFolder(dest)
	if err != nil {
		return "", err
	}

	var the7zPath2 string
	if the7zPath == nil {
		the7zPath2 = "7z"
	} else {
		the7zPath2 = *the7zPath
	}
		
	var logLevel string
	if os.Getenv("RUNNER_DEBUG") == "1" {
		logLevel = "-bb1"
	} else {
		logLevel = "-bb0"
	}
	cmd := exec.Command(the7zPath2, "x", logLevel, "-bd", "-sccUTF-8", file)
	cmd.Dir = dest2
	err = cmd.Run()
	if err != nil {
		return "", err
	}
	
	return dest2, nil
}

func createExtractFolder(dest *string) (string, error) {
	var dest2 string
	if dest == nil {
		dest2 = filepath.Join(os.Getenv("RUNNER_TEMP"), uuid.NewString())
	} else {
		dest2 = *dest
	}
	err := os.MkdirAll(dest2, 0o755)
	if err != nil {
		return "", err
	}
	return dest2, nil
}
