package filex

import "strings"

func GetSuffix(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}
