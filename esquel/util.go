package esquel

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"unicode"
	
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	paramMatcher = regexp.MustCompile(ParamPrefix + "[a-zA-Z0-9_]+")
)

var (
	escaper = strings.NewReplacer("`", "", "\"", "", "'", "")
)

func CreateTableStructure(fields []Field) string {
	result := make([]string, 0)
	for _, f := range fields {
		result = append(result, fmt.Sprintf("%s %s,", f.Name, f.Props))
	}
	result[len(result)-1] = strings.TrimSuffix(result[len(result)-1], ",")
	return strings.Join(result, "\n")
}

func MergeFields(a []Field, b []Field) []Field {
	result := make([]Field, 0)
	replacedFieldsNames := make([]string, 0)
	for i, af := range a {
		var exist bool
		var replacement Field
		for _, bf := range b {
			if af.Name == bf.Name {
				exist = true
				replacement = bf
				replacedFieldsNames = append(replacedFieldsNames, af.Name)
			}
		}
		if exist {
			a[i] = replacement
		}
		result = append(result, a[i])
	}
	for _, bf := range b {
		if slices.Contains(replacedFieldsNames, bf.Name) {
			continue
		}
		result = append(result, bf)
	}
	return result
}

func formatSql(q string) string {
	replacer := strings.NewReplacer("\t", " ", "\n", " ")
	q = replacer.Replace(q)
	q = strings.Join(strings.Fields(q), " ")
	q = strings.Replace(q, " ;", ";", -1)
	return q
}

func getSubstringIndexes(str, substr string) []int {
	result := make([]int, 0)
	scanned := 0
	for {
		i := strings.Index(str, substr)
		if i < 0 {
			break
		}
		if i > -1 {
			result = append(result, i+scanned)
			scanned += i + len(substr)
		}
		str = str[i+len(substr):]
	}
	return result
}

func createSlicePlaceholder(len int) string {
	result := make([]string, len)
	for i := 0; i < len; i++ {
		result[i] = Placeholder
	}
	return strings.Join(result, ",")
}

func replaceStringAtIndex(str, substr, newSubstr string, index int) string {
	return str[:index] + newSubstr + str[index+len(substr):]
}

func latinize(value string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, err := transform.String(t, value)
	if err != nil {
		return ""
	}
	return result
}

func replaceSpecialCharacters(value string) string {
	replacer := regexp.MustCompile(`[,-/#$%^&;{}=\-_~()@]`)
	marksReplacer := strings.NewReplacer("'", "’")
	value = replacer.ReplaceAllString(value, " ")
	value = marksReplacer.Replace(value)
	return value
}

func Normalize(value string) string {
	value = strings.ToLower(value)
	value = latinize(value)
	value = replaceSpecialCharacters(value)
	return value
}

func RemoveAccents(value string) string {
	return fmt.Sprintf("unaccent(lower(replace(%s, ' ', '-')))", value)
}

func Escape(value string) string {
	return escaper.Replace(value)
}
