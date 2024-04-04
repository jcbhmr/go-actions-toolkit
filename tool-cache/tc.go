package tc

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/jcbhmr/actions-toolkit.go/core"
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
	err := os.MkdirAll(dest2, 0o755)
	if err != nil {
		return "", err
	}
	core.Debug(fmt.Sprintf("downloading %s", url))
	core.Debug(fmt.Sprintf("destination %s", dest2))

	return "", nil
}
