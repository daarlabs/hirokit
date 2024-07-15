package devtool

import (
	"regexp"
	"strings"
	
	. "github.com/daarlabs/hirokit/gox"
	"github.com/daarlabs/hirokit/tempest"
)

var (
	sqlStringMatcher = regexp.MustCompile(`'.*'`)
)

func FormatSql(value string) string {
	value = strings.ReplaceAll(value, "SELECT", highlightSql("SELECT"))
	value = strings.ReplaceAll(value, "FROM", highlightSql("FROM"))
	value = strings.ReplaceAll(value, "INSERT INTO", highlightSql("INSERT INTO"))
	value = strings.ReplaceAll(value, "VALUES", highlightSql("VALUES"))
	value = strings.ReplaceAll(value, "UPDATE", highlightSql("UPDATE"))
	value = strings.ReplaceAll(value, " SET", highlightSql("SET"))
	value = strings.ReplaceAll(value, "DELETE", highlightSql("DELETE"))
	value = strings.ReplaceAll(value, "LEFT JOIN", highlightSql("LEFT JOIN"))
	value = strings.ReplaceAll(value, "RIGHT JOIN", highlightSql("RIGHT JOIN"))
	value = strings.ReplaceAll(value, "INNER JOIN", highlightSql("INNER JOIN"))
	value = strings.ReplaceAll(value, " AS ", highlightSql(" AS "))
	value = strings.ReplaceAll(value, " ON ", highlightSql(" ON "))
	value = strings.ReplaceAll(value, " IN ", highlightSql(" IN "))
	value = strings.ReplaceAll(value, " ANY", highlightSql(" ANY"))
	value = strings.ReplaceAll(value, " @@ ", highlightSql(" @@ "))
	value = strings.ReplaceAll(value, "WHERE", highlightSql("WHERE"))
	value = strings.ReplaceAll(value, "AND", highlightSql("AND"))
	value = strings.ReplaceAll(value, "GROUP BY", highlightSql("GROUP BY"))
	value = strings.ReplaceAll(value, "ORDER BY", highlightSql("ORDER BY"))
	value = strings.ReplaceAll(value, "OFFSET", highlightSql("OFFSET"))
	value = strings.ReplaceAll(value, "LIMIT", highlightSql("LIMIT"))
	for _, item := range sqlStringMatcher.FindAllString(value, -1) {
		value = strings.ReplaceAll(value, item, highlightSqlText(item))
	}
	return value
}

func highlightSql(value string) string {
	return Render(Span(tempest.Class().Name(DynamicStyle).TextOrange(400), Text(value)))
}
func highlightSqlText(value string) string {
	return Render(Span(tempest.Class().Name(DynamicStyle).TextFuchsia(400), Text(value)))
}
