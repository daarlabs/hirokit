package devtool

import (
	"strings"
	
	. "github.com/daarlabs/hirokit/gox"
	"github.com/daarlabs/hirokit/tempest"
)

func FormatSql(value string) string {
	value = strings.ReplaceAll(value, "SELECT", "\n"+highlightSql("SELECT"))
	value = strings.ReplaceAll(value, "FROM", "\n"+highlightSql("FROM"))
	value = strings.ReplaceAll(value, "INSERT INTO", "\n"+highlightSql("INSERT INTO"))
	value = strings.ReplaceAll(value, "VALUES", "\n"+highlightSql("VALUES"))
	value = strings.ReplaceAll(value, "UPDATE", "\n"+highlightSql("UPDATE"))
	value = strings.ReplaceAll(value, " SET", "\n"+highlightSql("SET"))
	value = strings.ReplaceAll(value, "DELETE", "\n"+highlightSql("DELETE"))
	value = strings.ReplaceAll(value, "LEFT JOIN", "\n"+highlightSql("LEFT JOIN"))
	value = strings.ReplaceAll(value, "RIGHT JOIN", "\n"+highlightSql("RIGHT JOIN"))
	value = strings.ReplaceAll(value, "INNER JOIN", "\n"+highlightSql("INNER JOIN"))
	value = strings.ReplaceAll(value, " AS ", highlightSql(" AS "))
	value = strings.ReplaceAll(value, " ON ", highlightSql(" ON "))
	value = strings.ReplaceAll(value, "WHERE", "\n"+highlightSql("WHERE"))
	value = strings.ReplaceAll(value, "AND", "\n\t"+highlightSql("AND"))
	value = strings.ReplaceAll(value, "GROUP BY", "\n"+highlightSql("GROUP BY"))
	value = strings.ReplaceAll(value, "ORDER BY", "\n"+highlightSql("ORDER BY"))
	value = strings.ReplaceAll(value, "OFFSET", "\n"+highlightSql("OFFSET"))
	value = strings.ReplaceAll(value, "LIMIT", "\n"+highlightSql("LIMIT"))
	value = strings.ReplaceAll(value, "(", "\n(\n\t")
	value = strings.ReplaceAll(value, ")", "\n)")
	return value
}

func highlightSql(value string) string {
	return Render(Span(tempest.Class().TextOrange(400), Text(value)))
}
