package config

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unicode"
)

const (
	defaultTag      = "default"
	mapstructureTag = "mapstructure"
)

var schemaCache sync.Map

type fieldSpec struct {
	key          string
	defaultValue string
	hasDefault   bool
}

func (field fieldSpec) env(prefix string) string {
	name := strings.ToUpper(strings.ReplaceAll(field.key, ".", "_"))
	return prefix + "_" + name
}

func schemaFor(value any) ([]fieldSpec, error) {
	valueType := reflect.TypeOf(value)
	for valueType.Kind() == reflect.Pointer {
		valueType = valueType.Elem()
	}
	if valueType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("configuration schema must be a struct")
	}
	if cached, ok := schemaCache.Load(valueType); ok {
		return cached.([]fieldSpec), nil
	}

	fields, err := collectFields(valueType, nil)
	if err != nil {
		return nil, err
	}
	schemaCache.Store(valueType, fields)
	return fields, nil
}

func collectFields(valueType reflect.Type, path []string) ([]fieldSpec, error) {
	fields := make([]fieldSpec, 0, valueType.NumField())
	for structField := range valueType.Fields() {
		structField := structField
		if structField.PkgPath != "" {
			continue
		}

		name, squash, skip := fieldTag(structField)
		if skip {
			continue
		}

		fieldType := structField.Type
		for fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
		}

		fieldPath := appendPath(path, name, squash)
		if fieldType.Kind() == reflect.Struct {
			nested, err := collectFields(fieldType, fieldPath)
			if err != nil {
				return nil, err
			}
			fields = append(fields, nested...)
			continue
		}

		key := strings.Join(fieldPath, ".")
		if key == "" {
			key = toSnake(structField.Name)
		}

		defaultValue, hasDefault := structField.Tag.Lookup(defaultTag)
		fields = append(fields, fieldSpec{
			key:          key,
			defaultValue: defaultValue,
			hasDefault:   hasDefault,
		})
	}

	return fields, nil
}

func fieldTag(structField reflect.StructField) (string, bool, bool) {
	tag := structField.Tag.Get(mapstructureTag)
	if tag == "-" {
		return "", false, true
	}
	if tag == "" {
		return toSnake(structField.Name), false, false
	}

	parts := strings.Split(tag, ",")
	name := parts[0]
	squash := false
	for _, part := range parts[1:] {
		if part == "squash" {
			squash = true
		}
	}
	return name, squash, false
}

func appendPath(path []string, name string, squash bool) []string {
	if squash || name == "" {
		return path
	}

	next := make([]string, 0, len(path)+1)
	next = append(next, path...)
	next = append(next, name)
	return next
}

func toSnake(value string) string {
	var output strings.Builder
	for index, current := range value {
		if index > 0 && unicode.IsUpper(current) {
			output.WriteRune('_')
		}
		output.WriteRune(unicode.ToLower(current))
	}
	return output.String()
}
