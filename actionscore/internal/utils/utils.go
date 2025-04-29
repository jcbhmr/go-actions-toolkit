package utils

import (
	"encoding/json"
	"reflect"
)

func ToCommandValue(input any) (string, error) {
	if input == nil {
		return "", nil
	}
	switch reflect.TypeOf(input).Kind() {
	case reflect.Pointer, reflect.UnsafePointer, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		if reflect.ValueOf(input).IsNil() {
			return "", nil
		}
	}
	if v, ok := input.(string); ok {
		return v, nil
	}
	bytes, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Keep in sync with core.AnnotationProperties
type InternalCoreAnnotationProperties struct {
	Title       *string
	File        *string
	StartLine   *string
	EndLine     *string
	StartColumn *string
	EndColumn   *string
}

type commandCommandProperties = map[string]any

func ToCommandProperties(annotationProperties InternalCoreAnnotationProperties) commandCommandProperties {
	if annotationProperties.Title == nil &&
		annotationProperties.File == nil &&
		annotationProperties.StartLine == nil &&
		annotationProperties.EndLine == nil &&
		annotationProperties.StartColumn == nil &&
		annotationProperties.EndColumn == nil {
		return commandCommandProperties{}
	}

	return commandCommandProperties{
		"title":       annotationProperties.Title,
		"file":        annotationProperties.File,
		"startLine":   annotationProperties.StartLine,
		"endLine":     annotationProperties.EndLine,
		"startColumn": annotationProperties.StartColumn,
		"endColumn":   annotationProperties.EndColumn,
	}
}
