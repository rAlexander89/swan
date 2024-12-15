package utils

import (
	"fmt"
	"strings"
	"unicode"
)

type Field struct {
	Name     string
	DataType string
}

func ToSnakeCase(s string) string {
	var result strings.Builder

	for i, r := range s {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}

	return result.String()
}

func ParseFields(fieldsStr string) ([]Field, error) {
	// remove brackets and split by comma
	fieldsStr = strings.TrimPrefix(fieldsStr, "[")
	fieldsStr = strings.TrimSuffix(fieldsStr, "]")

	if fieldsStr == "" {
		return nil, fmt.Errorf("no fields provided")
	}

	fieldPairs := strings.Split(fieldsStr, ",")
	fields := make([]Field, 0, len(fieldPairs))

	for _, pair := range fieldPairs {
		// trim spaces and split by space
		parts := strings.Split(strings.TrimSpace(pair), " ")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid field format: %s", pair)
		}

		fields = append(fields, Field{
			Name:     strings.TrimSpace(parts[0]),
			DataType: strings.TrimSpace(parts[1]),
		})
	}

	return fields, nil
}

func GenerateTags(fieldName string, tags []string) string {
	tagStr := ""
	for i, tag := range tags {
		if i > 0 {
			tagStr += " "
		}
		tagStr += fmt.Sprintf(`%s:"%s"`, tag, ToSnakeCase(fieldName))
	}
	return tagStr
}

func ParseTags(tagsStr string) ([]string, error) {
	// remove brackets and split by comma
	tagsStr = strings.TrimPrefix(tagsStr, "[")
	tagsStr = strings.TrimSuffix(tagsStr, "]")

	if tagsStr == "" {
		return nil, fmt.Errorf("no tags provided")
	}

	tags := strings.Split(tagsStr, ",")
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}

	return tags, nil
}

func ToUpperFirst(s string) string {
	if s == "" {
		return ""
	}

	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func SnakeToPascal(s string) string {
	var pascalCase string
	words := strings.Split(s, "_")

	for _, w := range words {
		pascalCase += ToUpperFirst(w)
	}

	return pascalCase
}

func PascalToLower(s string) string {
	// handle empty string case
	if len(s) == 0 {
		return s
	}

	var result strings.Builder

	// start with first character lowercase
	result.WriteString(strings.ToLower(string(s[1])))

	// iterate starting from second character
	for i := 1; i < len(s); i++ {
		if unicode.IsUpper(rune(s[i])) {
			result.WriteRune('_')
			result.WriteString(strings.ToLower(string(s[i])))
		} else {
			result.WriteRune(rune(s[i]))
		}
	}

	return result.String()
}
