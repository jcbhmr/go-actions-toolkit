package core

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"text/template"
)

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
	// Replace all existing content in summary file with buffer contents
	// (optional) default: false
	Overwrite *bool
}

// Writes text in the buffer to the summary buffer file and empties buffer. Will append by default.
//
// @param {SummaryWriteOptions} [options] (optional) options for write operation
//
// @returns {Promise<Summary>} summary instance
func (s *summary) Write(options *SummaryWriteOptions) (*summary, error) {
	var path string
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

// Clears the summary buffer and wipes the summary file
//
// @returns {Summary} summary instance
func (s *summary) Clear() (*summary, error) {
	s.buffer = ""
	_, err := s.Write(nil)
	if err != nil {
		return s, err
	}
	return s, nil
}

// Returns the current summary buffer as a string
//
// @returns {string} string of summary buffer
func (s *summary) Stringify() string {
	return s.buffer
}

// If the summary buffer is empty
//
// @returns {boolen} true if the buffer is empty
func (s *summary) IsEmptyBuffer() bool {
	return s.buffer == ""
}

// Resets the summary buffer without writing to summary file
//
// @returns {Summary} summary instance
func (s *summary) EmptyBuffer() *summary {
	s.buffer = ""
	return s
}

// Adds raw text to the summary buffer
//
// @param {string} text content to add
// @param {boolean} [addEOL=false] (optional) append an EOL to the raw text (default: false)
//
// @returns {Summary} summary instance
func (s *summary) AddRaw(text string, addEol *bool) *summary {
	s.buffer += text
	addEol2 := addEol != nil && *addEol
	if addEol2 {
		s.buffer += "\n"
	}
	return s
}

// Adds the operating system-specific end-of-line marker to the buffer
//
// @returns {Summary} summary instance
func (s *summary) AddEol() *summary {
	s.buffer += "\n"
	return s
}

// Adds an HTML codeblock to the summary buffer
//
// @param {string} code content to render within fenced code block
// @param {string} lang (optional) language to syntax highlight code
// @returns {Summary} summary instance
func (s *summary) AddCodeBlock(code string, lang *string) *summary {
	lang2 := ""
	if lang != nil {
		lang2 = *lang
	}
	s.buffer += fmt.Sprintf("```%s\n%s\n```\n", lang2, code)
	return s
}

// Adds an HTML list to the summary buffer
//
// @param {string[]} items list of items to render
// @param {boolean} [ordered=false] (optional) if the rendered list should be ordered or not (default: false)
//
// @returns {Summary} summary instance
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
	// Cell content
	Data string
	// Render cell as header
	// (optional) default: false
	Header *bool
	// Number of columns the cell extends
	// (optional) default: '1'
	Colspan *string
	// Number of rows the cell extends
	// (optional) default: '1'
	Rowspan *string
}

// Adds an HTML table to the summary buffer
//
// @param {SummaryTableCell[]} rows table rows
//
// @returns {Summary} summary instance
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

// Adds a collapsable HTML details element to the summary buffer
//
// @param {string} label text for the closed state
// @param {string} content collapsable content
//
// @returns {Summary} summary instance
func (s *summary) AddDetails(label string, content string) *summary {
	s.buffer += fmt.Sprintf("<details><summary>%s</summary>%s</details>\n", label, content)
	return s
}

type SummaryImageOptions struct {
	// The width of the image in pixels. Must be an integer without a unit.
	// (optional)
	width *string
	// The height of the image in pixels. Must be an integer without a unit.
	// (optional)
	height *string
}

// Adds an HTML image tag to the summary buffer
//
// @param {string} src path to the image you to embed
// @param {string} alt text description of the image
// @param {SummaryImageOptions} options (optional) addition image attributes
//
// @returns {Summary} summary instance
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

// Adds an HTML section heading element
//
// @param {string} text heading text
// @param {number | string} [level=1] (optional) the heading level, default: 1
//
// @returns {Summary} summary instance
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

// Adds an HTML thematic break (<hr>) to the summary buffer
//
// @returns {Summary} summary instance
func (s *summary) AddSeparator() *summary {
	s.buffer += "<hr>\n"
	return s
}

// Adds an HTML line break (<br>) to the summary buffer
//
// @returns {Summary} summary instance
func (s *summary) AddBreak() *summary {
	s.buffer += "<br>\n"
	return s
}

// Adds an HTML blockquote to the summary buffer
//
// @param {string} text quote text
// @param {string} cite (optional) citation url
//
// @returns {Summary} summary instance
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

// Adds an HTML anchor tag to the summary buffer
//
// @param {string} text link text/content
// @param {string} href hyperlink
//
// @returns {Summary} summary instance
func (s *summary) AddLink(text string, href string) *summary {
	s.buffer += fmt.Sprintf("<a href=\"%s\">%s</a>\n", href, text)
	return s
}

var Summary = newSummary()

// Deprecated: use `core.summary`
var MarkdownSummary = Summary
