package utils

import (
	"errors"
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

func PascalToSnake(s string) string {
	var result strings.Builder
	var prev rune

	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) && prev != '_' && !unicode.IsUpper(prev) {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
		prev = r
	}

	return result.String()
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
	for i := 0; i < len(s); i++ {
		if unicode.IsUpper(rune(s[i])) {
			result.WriteRune('_')
			result.WriteString(strings.ToLower(string(s[i])))
		} else {
			result.WriteRune(rune(s[i]))
		}
	}

	return result.String()
}

// ParseArgFields takes an array of args and returns fields until it hits a flag
func ParseArgFields(args []string, startIndex int) ([]Field, error) {
	var fields []Field
	i := startIndex

	// collect args until we hit a flag or end
	for ; i < len(args); i++ {
		// stop if we hit a flag
		if strings.HasPrefix(args[i], "-") {
			break
		}

		// need at least 2 more args for a field
		if i+1 >= len(args) {
			return nil, errors.New("incomplete field definition")
		}

		// clean any trailing commas
		name := strings.TrimSuffix(args[i], ",")
		dataType := strings.TrimSuffix(args[i+1], ",")

		fields = append(fields, Field{
			Name:     name,
			DataType: dataType,
		})

		// skip the type arg since we used it
		i++
	}

	if len(fields) == 0 {
		return nil, errors.New("no fields provided")
	}

	// return the last processed index so caller knows where to continue
	return fields, nil
}

func ParseArgTags(args []string, startIndex int) ([]string, error) {
	var tags []string
	i := startIndex

	for ; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-") {
			break
		}
		tag := strings.TrimSuffix(args[i], ",")
		tags = append(tags, tag)
	}

	if len(tags) == 0 {
		return nil, errors.New("no tags provided")
	}

	return tags, nil
}
