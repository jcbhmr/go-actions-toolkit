package core

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/google/uuid"
)

// Interface for getInput options
type InputOptions struct {
	// Optional. Whether the input is required. If required and not present, will throw. Defaults to false
	Required *bool
	// Optional. Whether leading/trailing whitespace will be trimmed for the input. Defaults to true
	TrimWhitespace *bool
}

// The code to exit an action
type ExitCode int

const (
	// A code indicating that the action was successful
	ExitCodeSuccess ExitCode = 0
	// A code indicating that the action was a failure
	ExitCodeFailure ExitCode = 1
)

// Optional properties that can be sent with annotation commands (notice, error, and warning)
// See: https://docs.github.com/en/rest/reference/checks#create-a-check-run for more information about annotations.
type AnnotationProperties struct {
	// A title for the annotation.
	Title *string
	// The path of the file for which the annotation should be created.
	File *string
	// The start line for the annotation.
	StartLine *int
	// The end line for the annotation. Defaults to `startLine` when `startLine` is provided.
	EndLine *int
	// The start column for the annotation. Cannot be sent when `startLine` and `endLine` are different values.
	StartColumn *int
	// The end column for the annotation. Cannot be sent when `startLine` and `endLine` are different values.
	// Defaults to `startColumn` when `startColumn` is provided.
	EndColumn *int
}

func encodeCommandProperty(property string) string {
	replacer := strings.NewReplacer(
		"%", "%25",
		"\r", "%0D",
		"\n", "%0A",
		":", "%3A",
		",", "%2C",
	)
	return replacer.Replace(property)
}

func encodeCommandData(data string) string {
	replacer := strings.NewReplacer(
		"%", "%25",
		"\r", "%0D",
		"\n", "%0A",
	)
	return replacer.Replace(data)
}

func toCommandString(value any) string {
	if value == nil {
		return ""
	}
	switch value := value.(type) {
	case string:
		return value
	default:
		bytes, err := json.Marshal(value)
		if err != nil {
			panic(err)
		}
		return string(bytes)
	}
}

// Sets env variable for this action and future actions in the job
// @param name the name of the variable to set
// @param val the value of the variable. Non-string values will be converted to a string via JSON.stringify
func ExportVariable(name string, value any) {
	valueString := toCommandString(value)
	githubEnv, ok := os.LookupEnv("GITHUB_ENV")
	if ok {
		file, err := os.OpenFile(githubEnv, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		delimiter := uuid.NewString()
		_, err = file.WriteString(fmt.Sprintf("%s<<%s\n%s\n%s", name, delimiter, valueString, delimiter))
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Printf("::set-env name=%s::%s\n", encodeCommandProperty(name), encodeCommandData(valueString))
	}
}

// Registers a secret which will get masked from logs
// @param secret value of the secret
func SetSecret(secret string) {
	fmt.Printf("::add-mask::%s\n", encodeCommandData(secret))
}

// Prepends inputPath to the PATH (for this action and future actions)
// @param inputPath
func AddPath(inputPath string) {
	githubPath, ok := os.LookupEnv("GITHUB_PATH")
	if ok {
		file, err := os.OpenFile(githubPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		_, err = file.WriteString(inputPath + "\n")
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Printf("::add-path::%s\n", encodeCommandData(inputPath))
	}
	os.Setenv("PATH", fmt.Sprintf("%s%s%s", inputPath, string(os.PathListSeparator), os.Getenv("PATH")))
}

// Gets the value of an input.
// Unless trimWhitespace is set to false in InputOptions, the value is also trimmed.
// Returns an empty string if the value is not defined.
//
// @param     name     name of the input to get
// @param     options  optional. See InputOptions.
// @returns   string
func GetInput(name string, options *InputOptions) (string, error) {
	value := os.Getenv(fmt.Sprintf("INPUT_%s", strings.ReplaceAll(strings.ToUpper(name), " ", "_")))
	required := options != nil && options.Required != nil && *options.Required
	if required && value == "" {
		return "", fmt.Errorf("input required and not supplied: %s", name)
	}
	trimWhitespace := options == nil || options.TrimWhitespace == nil || *options.TrimWhitespace
	if trimWhitespace {
		value = strings.TrimSpace(value)
	}
	return value, nil
}

// Gets the values of an multiline input.  Each value is also trimmed.
//
// @param     name     name of the input to get
// @param     options  optional. See InputOptions.
// @returns   string[]
func GetMultilineInput(name string, options *InputOptions) ([]string, error) {
	value, err := GetInput(name, options)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(value, "\n")
	lines2 := []string{}
	for _, line := range lines {
		if line != "" {
			lines2 = append(lines2, line)
		}
	}
	trimWhitespace := options == nil || options.TrimWhitespace == nil || *options.TrimWhitespace
	if trimWhitespace {
		for i, line := range lines2 {
			lines2[i] = strings.TrimSpace(line)
		}
	}
	return lines2, nil
}

// Gets the input value of the boolean type in the YAML 1.2 "core schema" specification.
// Support boolean input list: `true | True | TRUE | false | False | FALSE` .
// The return value is also in boolean type.
// ref: https://yaml.org/spec/1.2/spec.html#id2804923
//
// @param     name     name of the input to get
// @param     options  optional. See InputOptions.
// @returns   boolean
func GetBooleanInput(name string, options *InputOptions) (bool, error) {
	value, err := GetInput(name, options)
	if err != nil {
		return false, err
	}
	if slices.Contains([]string{"true", "True", "TRUE"}, value) {
		return true, nil
	}
	if slices.Contains([]string{"false", "False", "FALSE"}, value) {
		return false, nil
	}
	return false, fmt.Errorf("input does not meet YAML 1.2 \"Core Schema\" specification: %s\nSupport boolean input list: `true | True | TRUE | false | False | FALSE`", value)
}

// Sets the value of an output.
//
// @param     name     name of the output to set
// @param     value    value to store. Non-string values will be converted to a string via JSON.stringify
func SetOutput(name string, value any) {
	valueString := toCommandString(value)
	githubOutput, ok := os.LookupEnv("GITHUB_OUTPUT")
	if ok {
		file, err := os.OpenFile(githubOutput, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		delimiter := uuid.NewString()
		_, err = file.WriteString(fmt.Sprintf("%s<<%s\n%s\n%s", name, delimiter, valueString, delimiter))
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Printf("::set-output name=%s::%s\n", encodeCommandProperty(name), encodeCommandData(valueString))
	}
}

// Enables or disables the echoing of commands into stdout for the rest of the step.
// Echoing is disabled by default if ACTIONS_STEP_DEBUG is not set.
func SetCommandEcho(enabled bool) {
	if enabled {
		fmt.Println("::echo::on")
	} else {
		fmt.Println("::echo::off")
	}
}

// Sets the action status to failed.
// When the action exits it will be with an exit code of 1
// @param message add error issue message
func SetFailed(message any) {
	Error(message, nil)
	os.Exit(1)
}

// Gets whether Actions Step Debug is on or not
func IsDebug() bool {
	return os.Getenv("RUNNER_DEBUG") == "1"
}

// Writes debug message to user log
// @param message debug message
func Debug(message any) {
	messageString := toCommandString(message)
	fmt.Printf("::debug::%s\n", encodeCommandData(messageString))
}

func annotationPropertiesToCommandPropertiesString(properties AnnotationProperties) string {
	parts := []string{}
	if properties.Title != nil {
		parts = append(parts, fmt.Sprintf("title=%s", encodeCommandProperty(*properties.Title)))
	}
	if properties.File != nil {
		parts = append(parts, fmt.Sprintf("file=%s", encodeCommandProperty(*properties.File)))
	}
	if properties.StartLine != nil {
		parts = append(parts, fmt.Sprintf("line=%d", *properties.StartLine))
	}
	if properties.EndLine != nil {
		parts = append(parts, fmt.Sprintf("endLine=%d", *properties.EndLine))
	}
	if properties.StartColumn != nil {
		parts = append(parts, fmt.Sprintf("col=%d", *properties.StartColumn))
	}
	if properties.EndColumn != nil {
		parts = append(parts, fmt.Sprintf("endCol=%d", *properties.EndColumn))
	}
	return strings.Join(parts, ",")
}

// Adds an error issue
// @param message error issue message. Errors will be converted to string via toString()
// @param properties optional properties to add to the annotation.
func Error(message any, properties *AnnotationProperties) {
	commandProperties := ""
	if properties != nil {
		commandProperties = annotationPropertiesToCommandPropertiesString(*properties)
		if commandProperties != "" {
			commandProperties = " " + commandProperties
		}
	}
	switch message := message.(type) {
	case string:
		fmt.Printf("::error%s::%s\n", commandProperties, encodeCommandData(message))
	case error:
		fmt.Printf("::error%s::%s\n", commandProperties, encodeCommandData(message.Error()))
	default:
		panic("unexpected type")
	}
}

// Adds a warning issue
// @param message warning issue message. Errors will be converted to string via toString()
// @param properties optional properties to add to the annotation.
func Warning(message any, properties *AnnotationProperties) {
	commandProperties := ""
	if properties != nil {
		commandProperties = annotationPropertiesToCommandPropertiesString(*properties)
		if commandProperties != "" {
			commandProperties = " " + commandProperties
		}
	}
	switch message := message.(type) {
	case string:
		fmt.Printf("::warning%s::%s\n", commandProperties, encodeCommandData(message))
	case error:
		fmt.Printf("::warning%s::%s\n", commandProperties, encodeCommandData(message.Error()))
	default:
		panic("unexpected type")
	}
}

// Adds a notice issue
// @param message notice issue message. Errors will be converted to string via toString()
// @param properties optional properties to add to the annotation.
func Notice(message any, properties *AnnotationProperties) {
	commandProperties := ""
	if properties != nil {
		commandProperties = annotationPropertiesToCommandPropertiesString(*properties)
		if commandProperties != "" {
			commandProperties = " " + commandProperties
		}
	}
	switch message := message.(type) {
	case string:
		fmt.Printf("::notice%s::%s\n", commandProperties, encodeCommandData(message))
	case error:
		fmt.Printf("::notice%s::%s\n", commandProperties, encodeCommandData(message.Error()))
	default:
		panic("unexpected type")
	}
}

// Writes info to log with console.log.
// @param message info message
func Info(message string) {
	fmt.Println(message)
}

// Begin an output group.
// Output until the next `groupEnd` will be foldable in this group
//
// @param name The name of the output group
func StartGroup(name string) {
	fmt.Printf("::group::%s\n", encodeCommandData(name))
}

// End an output group.
func EndGroup() {
	fmt.Println("::endgroup::")
}

// Wrap an asynchronous function call in a group.
//
// Returns the same type as the function itself.
//
// @param name The name of the group
// @param fn The function to wrap in the group
func Group(name string, fn func()) {
	StartGroup(name)
	defer EndGroup()
	fn()
}

func Group1[T any](name string, fn func() T) T {
	StartGroup(name)
	defer EndGroup()
	return fn()
}

func Group2[T1 any, T2 any](name string, fn func() (T1, T2)) (T1, T2) {
	StartGroup(name)
	defer EndGroup()
	return fn()
}

// Saves state for current action, the state can only be retrieved by this action's post job execution.
//
// @param     name     name of the state to store
// @param     value    value to store. Non-string values will be converted to a string via JSON.stringify
func SaveState(name string, value any) {
	valueString := toCommandString(value)
	githubState, ok := os.LookupEnv("GITHUB_STATE")
	if ok {
		file, err := os.OpenFile(githubState, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		delimiter := uuid.NewString()
		_, err = file.WriteString(fmt.Sprintf("%s<<%s\n%s\n%s", name, delimiter, valueString, delimiter))
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Printf("::save-state name=%s::%s\n", encodeCommandProperty(name), encodeCommandData(valueString))
	}
}

// Gets the value of an state set by this action's main execution.
//
// @param     name     name of the state to get
// @returns   string
func GetState(name string) string {
	return os.Getenv(fmt.Sprintf("STATE_%s", name))
}
