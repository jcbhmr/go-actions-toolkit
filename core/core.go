package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strings"
	"text/template"

	"github.com/google/uuid"
)

type InputOptions struct {
	Required       *bool
	TrimWhitespace *bool
}

type ExitCode int

const (
	Success ExitCode = 0
	Failure ExitCode = 1
)

type AnnotationProperties struct {
	Title       *string
	File        *string
	StartLine   *int
	EndLine     *int
	StartColumn *int
	EndColumn   *int
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
	switch value := value.(type) {
	case nil:
		return ""
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

func SetSecret(secret string) {
	fmt.Printf("::add-mask::%s\n", encodeCommandData(secret))
}

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

func SetCommandEcho(enabled bool) {
	if enabled {
		fmt.Println("::echo::on")
	} else {
		fmt.Println("::echo::off")
	}
}

func SetFailed(message any) {
	Error(message, nil)
	os.Exit(1)
}

func IsDebug() bool {
	return os.Getenv("RUNNER_DEBUG") == "1"
}

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

func Info(message string) {
	fmt.Println(message)
}

func StartGroup(name string) {
	fmt.Printf("::group::%s\n", encodeCommandData(name))
}

func EndGroup() {
	fmt.Println("::endgroup::")
}

func Group(name string, fn func()) {
	StartGroup(name)
	defer EndGroup()
	fn()
}

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

func GetState(name string) string {
	return os.Getenv(fmt.Sprintf("STATE_%s", name))
}

type tokenResponse struct {
	Value string
}

func GetIdToken(audience *string) (string, error) {
	url, ok := os.LookupEnv("ACTIONS_ID_TOKEN_REQUEST_URL")
	if !ok {
		return "", errors.New("unable to get ACTIONS_ID_TOKEN_REQUEST_URL env variable")
	}
	if audience != nil {
		url += fmt.Sprintf("&audience=%s", *audience)
	}
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	var target tokenResponse
	err = json.NewDecoder(res.Body).Decode(&target)
	if err != nil {
		return "", err
	}
	return target.Value, nil
}

func ToPosixPath(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

func ToWin32Path(path string) string {
	return strings.ReplaceAll(path, "/", "\\")
}

func ToPlatformPath(path string) string {
	if runtime.GOOS == "windows" {
		return ToWin32Path(path)
	} else {
		return ToPosixPath(path)
	}
}

type summary struct {
	buffer string
	path   *string
}

func newSummary() *summary {
	return &summary{
		buffer: "",
		path:   nil,
	}
}

type SummaryWriteOptions struct {
	Overwrite *bool
}

func (s *summary) Write(options *SummaryWriteOptions) (*summary, error) {
	path := ""
	if s.path == nil {
		path = os.Getenv("GITHUB_STEP_SUMMARY")
		s.path = &path
	} else {
		path = *s.path
	}
	overwrite := options != nil && options.Overwrite != nil && *options.Overwrite
	if overwrite {
		err := os.WriteFile(path, []byte(s.buffer), 0644)
		if err != nil {
			return s, err
		}
	} else {
		file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return s, err
		}
		defer file.Close()
		_, err = file.WriteString(s.buffer)
		if err != nil {
			return s, err
		}
	}
	return s, nil
}

func (s *summary) Clear() (*summary, error) {
	s.buffer = ""
	_, err := s.Write(nil)
	if err != nil {
		return s, err
	}
	return s, nil
}

func (s *summary) Stringify() string {
	return s.buffer
}

func (s *summary) IsEmptyBuffer() bool {
	return s.buffer == ""
}

func (s *summary) EmptyBuffer() *summary {
	s.buffer = ""
	return s
}

func (s *summary) AddRaw(text string, addEol *bool) *summary {
	s.buffer += text
	addEol2 := addEol != nil && *addEol
	if addEol2 {
		s.buffer += "\n"
	}
	return s
}

func (s *summary) AddEol() *summary {
	s.buffer += "\n"
	return s
}

func (s *summary) AddCodeBlock(code string, lang *string) *summary {
	lang2 := ""
	if lang != nil {
		lang2 = *lang
	}
	s.buffer += fmt.Sprintf("```%s\n%s\n```\n", lang2, code)
	return s
}

func (s *summary) AddList(items []string, ordered *bool) *summary {
	ordered2 := ordered != nil && *ordered
	if ordered2 {
		tmpl, err := template.New("list").Parse("<ol>{{range .}}<li>{{.}}</li>{{end}}</ol>")
		if err != nil {
			panic(err)
		}
		sb := strings.Builder{}
		err = tmpl.Execute(&sb, items)
		if err != nil {
			panic(err)
		}
		s.buffer += sb.String()
		s.buffer += "\n"
		return s
	} else {
		tmpl, err := template.New("list").Parse("<ul>{{range .}}<li>{{.}}</li>{{end}}</ul>")
		if err != nil {
			panic(err)
		}
		sb := strings.Builder{}
		err = tmpl.Execute(&sb, items)
		if err != nil {
			panic(err)
		}
		s.buffer += sb.String()
		s.buffer += "\n"
		return s
	}
}

type SummaryTableCell struct {
	Data    string
	Header  *bool
	Colspan *string
	Rowspan *string
}

// Adds a `<table>` to the summary with the given rows. The `rows` parameter
// is a `(string | SummaryTableCell)[]`. Each row is either a string or an
// object with Data, Header, Colspan, and Rowspan properties.
func (s *summary) AddTable(rows []any) *summary {
	tmpl, err := template.New("table").Parse("<table>{{range .}}<tr>{{range .}}<td{{if .Header}} header{{end}}{{if .Colspan}} colspan={{.Colspan}}{{end}}{{if .Rowspan}} rowspan={{.Rowspan}}{{end}}>{{.Data}}</td>{{end}}</tr>{{end}}</table>")
	if err != nil {
		panic(err)
	}
	sb := strings.Builder{}
	err = tmpl.Execute(&sb, rows)
	if err != nil {
		panic(err)
	}
	s.buffer += sb.String()
	s.buffer += "\n"
	return s
}

func (s *summary) AddDetails(label string, content string) *summary {
	s.buffer += fmt.Sprintf("<details><summary>%s</summary>%s</details>\n", label, content)
	return s
}

type SummaryImageOptions struct {
	width  *string
	height *string
}

func (s *summary) AddImage(src string, alt string, options *SummaryImageOptions) *summary {
	tmpl, err := template.New("image").Parse("<img src=\"{{.Src}}\" alt=\"{{.Alt}}\"{{if .Width}} width=\"{{.Width}}\"{{end}}{{if .Height}} height=\"{{.Height}}\"{{end}}>\n")
	if err != nil {
		panic(err)
	}
	sb := strings.Builder{}
	width := ""
	if options != nil && options.width != nil {
		width = *options.width
	}
	height := ""
	if options != nil && options.height != nil {
		height = *options.height
	}
	err = tmpl.Execute(&sb, map[string]string{
		"Src":    src,
		"Alt":    alt,
		"Width":  width,
		"Height": height,
	})
	if err != nil {
		panic(err)
	}
	s.buffer += sb.String()
	return s
}

func (s *summary) AddHeading(text string, level any) *summary {
	hTag := "h1"
	switch level := level.(type) {
	case int:
		if 1 <= level && level <= 6 {
			hTag = fmt.Sprintf("h%d", level)
		}
	case string:
		if slices.Contains([]string{"1", "2", "3", "4", "5", "6"}, level) {
			hTag = fmt.Sprintf("h%s", level)
		}
	default:
		panic("unexpected type")
	}
	s.buffer += fmt.Sprintf("<%s>%s</%s>\n", hTag, text, hTag)
	return s
}

func (s *summary) AddSeparator() *summary {
	s.buffer += "<hr>\n"
	return s
}

func (s *summary) AddBreak() *summary {
	s.buffer += "<br>\n"
	return s
}

func (s *summary) AddQuote(text string, cite *string) *summary {
	tmpl, err := template.New("quote").Parse("<blockquote {{if .Cite}}cite=\"{{.Cite}}\"{{end}}>{{.Text}}</blockquote>")
	if err != nil {
		panic(err)
	}
	sb := strings.Builder{}
	cite2 := ""
	if cite != nil {
		cite2 = *cite
	}
	err = tmpl.Execute(&sb, map[string]string{
		"Text": text,
		"Cite": cite2,
	})
	if err != nil {
		panic(err)
	}
	s.buffer += sb.String()
	s.buffer += "\n"
	return s
}

func (s *summary) AddLink(text string, href string) *summary {
	s.buffer += fmt.Sprintf("<a href=\"%s\">%s</a>\n", href, text)
	return s
}

var Summary = newSummary()
var MarkdownSummary = Summary

var Platform = platform{
	Platform:  runtime.GOOS,
	Arch:      runtime.GOARCH,
	IsWindows: runtime.GOOS == "windows",
	IsMacOs:   runtime.GOOS == "darwin",
	IsLinux:   runtime.GOOS == "linux",
}

type platform struct {
	Platform  string
	Arch      string
	IsWindows bool
	IsMacOs   bool
	IsLinux   bool
}

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

func (p *platform) GetDetails() (platformDetails, error) {
	var name string
	var version string
	var err error
	if Platform.IsWindows {
		name, version, err = platformGetWindowsInfo()
	} else if Platform.IsMacOs {
		name, version, err = platformGetMacOsInfo()
	} else if Platform.IsLinux {
		name, version, err = platformGetLinuxInfo()
	}
	if err != nil {
		return platformDetails{}, err
	}
	details := platformDetails{
		Name:      name,
		Version:   version,
		Platform:  Platform.Platform,
		Arch:      Platform.Arch,
		IsWindows: Platform.IsWindows,
		IsMacOs:   Platform.IsMacOs,
		IsLinux:   Platform.IsLinux,
	}
	return details, nil
}
