package core

type InputOptions struct {
	Required       *bool
	TrimWhitespace *bool
}

type ExitCode int

const (
	ExitCodeSuccess ExitCode = 0
	ExitCodeFailure ExitCode = 1
)

type AnnotationProperties struct {
	Title       *string
	File        *string
	StartLine   *string
	EndLine     *string
	StartColumn *string
	EndColumn   *string
}

func ExportVariable(name string, val any) error {
	return exportVariable(name, val)
}

func SetSecret(secret string) error {
	return setSecret(secret)
}