package esquel

import (
	"fmt"
	"strings"
)

func CreateInsert(data map[string]any, modifier ...map[string]string) (string, string) {
	n := len(data)
	modifierExist := len(modifier) > 0
	columns := make([]string, n)
	placeholders := make([]string, n)
	i := 0
	for k := range data {
		columns[i] = k
		if !modifierExist {
			placeholders[i] = "@" + k
		}
		if modifierExist {
			m, ok := modifier[0][k]
			if !ok {
				placeholders[i] = "@" + k
			}
			if ok {
				placeholders[i] = fmt.Sprintf(m, "@"+k)
			}
		}
		i++
	}
	return strings.Join(columns, ", "), strings.Join(placeholders, ", ")
}

func CreateUpdate(data map[string]any, modifier ...map[string]string) string {
	n := len(data)
	modifierExist := len(modifier) > 0
	result := make([]string, n)
	i := 0
	for k := range data {
		if !modifierExist {
			result[i] = fmt.Sprintf("%[1]s = @%[1]s", k)
		}
		if modifierExist {
			m, ok := modifier[0][k]
			if !ok {
				result[i] = fmt.Sprintf("%[1]s = @%[1]s", k)
			}
			if ok {
				result[i] = fmt.Sprintf("%[1]s = %[2]s", k, fmt.Sprintf(m, "@"+k))
			}
		}
		i++
	}
	return strings.Join(append(result, "updated_at = CURRENT_TIMESTAMP"), ", ")
}
