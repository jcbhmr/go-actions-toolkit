package summary

import (
	"fmt"
	"os"
	"runtime"
)

const SummaryEnvVar = "GITHUB_STEP_SUMMARY"
const SummaryDocsURL = "https://docs.github.com/actions/using-workflows/workflow-commands-for-github-actions#adding-a-job-summary"

// [](SummaryTableCell | string)
type SummaryTableRow = []any

type SummaryTableCell struct {
	Data    string
	Header  *bool
	Colspan *string
	Rowspan *string
}

type SummaryImageOptions struct {
	Width  *string
	Height *string
}

type SummaryWriteOptions struct {
	Overwrite *bool
}

type Summary struct {
	buffer        string
	filePathValue *string
}

func NewSummary() *Summary {
	return &Summary{buffer: ""}
}

func (s *Summary) filePath() (string, error) {
	if s.filePathValue != nil {
		return *s.filePathValue, nil
	}

	pathFromEnv := os.Getenv(SummaryEnvVar)
	if pathFromEnv == "" {
		return "", fmt.Errorf("unable to find environment variable for %q. check if your runtime environment supports job summaries", SummaryEnvVar)
	}

	stats, err := os.Stat(pathFromEnv)
	if err == nil {
		if stats.Mode()&os.ModePerm != 0666 {
			err = fmt.Errorf("file %q is not writable", pathFromEnv)
		}
	}
	if err != nil {
		return "", fmt.Errorf("unable to access summary file %q. check if the file has correact read/write permissions", pathFromEnv)
	}

	s.filePathValue = &pathFromEnv
	return pathFromEnv, nil
}

func (s *Summary) wrap(tag string, content *string, attrs map[string]string) string {
	var htmlAttrs string
	for k, v := range attrs {
		htmlAttrs += fmt.Sprintf(` %s="%s"`, k, v)
	}

	if content == nil || *content == "" {
		return fmt.Sprintf("<%s%s/>", tag, htmlAttrs)
	}
	return fmt.Sprintf("<%s%s>%s</%[1]s>", tag, htmlAttrs, *content)
}

func (s *Summary) Write(options SummaryWriteOptions) (*Summary, error) {
	var overwrite bool
	if options.Overwrite != nil {
		overwrite = *options.Overwrite
	}
	filePath, err := s.filePath()
	if err != nil {
		return nil, err
	}
	var writeFunc func(string, []byte) error
	if overwrite {
		writeFunc = func(s string, b []byte) error {
			return os.WriteFile(s, b, 0666)
		}
	} else {
		writeFunc = func(s string, b []byte) error {
			f, err := os.OpenFile(s, os.O_APPEND|os.O_WRONLY, 0666)
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = f.Write(b)
			return err
		}
	}
	err = writeFunc(filePath, []byte(s.buffer))
	if err != nil {
		return nil, err
	}
	return s.EmptyBuffer(), nil
}

func ptr[T any](v T) *T {
	return &v
}

func (s *Summary) Clear() (*Summary, error) {
	return s.EmptyBuffer().Write(SummaryWriteOptions{Overwrite: ptr(true)})
}

func (s *Summary) Stringify() string {
	return s.buffer
}

func (s *Summary) IsEmptyBuffer() bool {
	return s.buffer == ""
}

func (s *Summary) EmptyBuffer() *Summary {
	s.buffer = ""
	return s
}

func (s *Summary) AddRaw(text string, addEOLRaw *bool) *Summary {
	var addEOL bool
	if addEOLRaw != nil {
		addEOL = *addEOLRaw
	}
	s.buffer += text
	if addEOL {
		return s.AddEOL()
	} else {
		return s
	}
}

var eol = func() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	} else {
		return "\n"
	}
}()

func (s *Summary) AddEOL() *Summary {
	s.buffer += eol
	return s
}

