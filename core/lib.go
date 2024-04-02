package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
)

// Interface for getInput options
type InputOptions struct {
	Required       bool `json:"required,omitempty"`
	TrimWhitespace bool `json:"trimWhitespace,omitempty"`
}

// The code to exit an action
type ExitCode int

const (
	Success ExitCode = iota
	Failure
)

// Optional properties that can be sent with annotation commands (notice, error, and warning)
type AnnotationProperties struct {
	Title       string `json:"title,omitempty"`
	File        string `json:"file,omitempty"`
	StartLine   int    `json:"startLine,omitempty"`
	EndLine     int    `json:"endLine,omitempty"`
	StartColumn int    `json:"startColumn,omitempty"`
	EndColumn   int    `json:"endColumn,omitempty"`
}

// Sets env variable for this action and future actions in the job
func ExportVariable(name string, val interface{}) {
	convertedVal, _ := json.Marshal(val)
	os.Setenv(name, string(convertedVal))

	filePath := os.Getenv("GITHUB_ENV")
	if filePath != "" {
		IssueFileCommand("ENV", prepareKeyValueMessage(name, val))
		return
	}

	IssueCommand("set-env", map[string]interface{}{"name": name}, string(convertedVal))
}

// Registers a secret which will get masked from logs
func SetSecret(secret string) {
	IssueCommand("add-mask", nil, secret)
}

// Prepends inputPath to the PATH (for this action and future actions)
func AddPath(inputPath string) {
	filePath := os.Getenv("GITHUB_PATH")
	if filePath != "" {
		IssueFileCommand("PATH", inputPath)
	} else {
		IssueCommand("add-path", nil, inputPath)
	}
	os.Setenv("PATH", fmt.Sprintf("%s%s%s", inputPath, string(os.PathListSeparator), os.Getenv("PATH")))
}

// Gets the value of an input.
// Unless trimWhitespace is set to false in InputOptions, the value is also trimmed.
// Returns an empty string if the value is not defined.
func GetInput(name string, options *InputOptions) string {
	val := os.Getenv(fmt.Sprintf("INPUT_%s", strings.ToUpper(strings.ReplaceAll(name, " ", "_"))))

	if options != nil && options.Required && val == "" {
		panic(fmt.Sprintf("Input required and not supplied: %s", name))
	}

	if options != nil && !options.TrimWhitespace {
		return val
	}

	return strings.TrimSpace(val)
}

// Gets the values of an multiline input.  Each value is also trimmed.
func GetMultilineInput(name string, options *InputOptions) []string {
	inputs := strings.Split(GetInput(name, options), "\n")
	var result []string
	for _, input := range inputs {
		if input != "" {
			if options != nil && !options.TrimWhitespace {
				result = append(result, input)
			} else {
				result = append(result, strings.TrimSpace(input))
			}
		}
	}
	return result
}

// Gets the input value of the boolean type in the YAML 1.2 "core schema" specification.
// Support boolean input list: `true | True | TRUE | false | False | FALSE` .
// The return value is also in boolean type.
func GetBooleanInput(name string, options *InputOptions) bool {
	trueValue := []string{"true", "True", "TRUE"}
	falseValue := []string{"false", "False", "FALSE"}
	val := GetInput(name, options)
	if contains(trueValue, val) {
		return true
	}
	if contains(falseValue, val) {
		return false
	}
	panic(fmt.Sprintf("Input does not meet YAML 1.2 \"Core Schema\" specification: %s\nSupport boolean input list: `true | True | TRUE | false | False | FALSE`", name))
}

// Sets the value of an output.
func SetOutput(name string, value interface{}) {
	filePath := os.Getenv("GITHUB_OUTPUT")
	if filePath != "" {
		IssueFileCommand("OUTPUT", prepareKeyValueMessage(name, value))
	} else {
		fmt.Println()
		IssueCommand("set-output", map[string]interface{}{"name": name}, toCommandValue(value))
	}
}

// Enables or disables the echoing of commands into stdout for the rest of the step.
// Echoing is disabled by default if ACTIONS_STEP_DEBUG is not set.
func SetCommandEcho(enabled bool) {
	Issue("echo", map[string]interface{}{"enabled": enabled})
}

// Sets the action status to failed.
// When the action exits it will be with an exit code of 1
func SetFailed(message string) {
	os.Exit(int(Failure))
	Error(message)
}

// Gets whether Actions Step Debug is on or not
func IsDebug() bool {
	return os.Getenv("RUNNER_DEBUG") == "1"
}

// Writes debug message to user log
func Debug(message string) {
	IssueCommand("debug", nil, message)
}

// Adds an error issue
func Error(message string, properties AnnotationProperties) {
	IssueCommand("error", toCommandProperties(properties), message)
}

// Adds a warning issue
func Warning(message string, properties AnnotationProperties) {
	IssueCommand("warning", toCommandProperties(properties), message)
}

// Adds a notice issue
func Notice(message string, properties AnnotationProperties) {
	IssueCommand("notice", toCommandProperties(properties), message)
}

// Writes info to log with console.log.
func Info(message string) {
	fmt.Println(message)
}

// Begin an output group.
func StartGroup(name string) {
	Issue("group", name)
}

// End an output group.
func EndGroup() {
	Issue("endgroup", nil)
}

// Wrap an asynchronous function call in a group.
func Group(name string, fn func() error) error {
	StartGroup(name)
	defer EndGroup()

	return fn()
}

// Saves state for current action, the state can only be retrieved by this action's post job execution.
func SaveState(name string, value interface{}) {
	filePath := os.Getenv("GITHUB_STATE")
	if filePath != "" {
		IssueFileCommand("STATE", prepareKeyValueMessage(name, value))
	} else {
		IssueCommand("save-state", map[string]interface{}{"name": name}, toCommandValue(value))
	}
}

// Gets the value of an state set by this action's main execution.
func GetState(name string) string {
	return os.Getenv(fmt.Sprintf("STATE_%s", name))
}

func GetIDToken(aud string) (string, error) {
	return OidcClient.GetIDToken(aud)
}

func contains(arr []string, val string) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}

func toCommandValue(val interface{}) string {
	switch v := val.(type) {
	case string:
		return v
	default:
		bytes, _ := json.Marshal(v)
		return string(bytes)
	}
}

func toCommandProperties(properties AnnotationProperties) map[string]interface{} {
	result := make(map[string]interface{})
	if properties.Title != "" {
		result["title"] = properties.Title
	}
	if properties.File != "" {
		result["file"] = properties.File
	}
	if properties.StartLine != 0 {
		result["startLine"] = properties.StartLine
	}
	if properties.EndLine != 0 {
		result["endLine"] = properties.EndLine
	}
	if properties.StartColumn != 0 {
		result["startColumn"] = properties.StartColumn
	}
	if properties.EndColumn != 0 {
		result["endColumn"] = properties.EndColumn
	}
	return result
}

func IssueCommand(command string, properties map[string]interface{}, message string) {
	cmd := map[string]interface{}{
		"command": command,
		"message": message,
	}
	if properties != nil {
		for k, v := range properties {
			cmd[k] = v
		}
	}
	bytes, _ := json.Marshal(cmd)
	fmt.Println(string(bytes))
}

func IssueFileCommand(command string, filePath string) {
	cmd := map[string]interface{}{
		"command": command,
		"message": filePath,
	}
	bytes, _ := json.Marshal(cmd)
	fmt.Println(string(bytes))
}

func Issue(command string, properties map[string]interface{}) {
	cmd := map[string]interface{}{
		"command": command,
	}
	if properties != nil {
		for k, v := range properties {
			cmd[k] = v
		}
	}
	bytes, _ := json.Marshal(cmd)
	fmt.Println(string(bytes))
}

func prepareKeyValueMessage(key string, value interface{}) string {
	return fmt.Sprintf("%s=%s", key, value)
}

func main() {
	// Your code here
}
