package dyna

import (
	"fmt"
	"strings"
	
	"github.com/daarlabs/hirokit/esquel"
	"github.com/daarlabs/hirokit/hiro"
	"github.com/daarlabs/hirokit/util/strx"
)

type Param struct {
	Fulltext string
	Offset   int
	Limit    int
	Order    []string
	Group    []string
	Fields   Fields
}

type OrderParam struct {
	Field     string
	Direction string
}

type Fields struct {
	Fulltext []string
	Order    map[string]string
}

func (p Param) Parse(c hiro.Ctx) Param {
	c.Parse().MustQuery(Fulltext, &p.Fulltext)
	c.Parse().MustQuery(Offset, &p.Offset)
	c.Parse().MustQuery(Limit, &p.Limit)
	c.Parse().Multiple().MustQuery(Order, &p.Order)
	return p
}

func (p Param) Use(q *esquel.Esquel) {
	useFulltextParam(q, p.Fulltext, p.Fields.Fulltext)
	useGroupParam(q, p.Group)
	useOrderParam(q, p.Order, p.Fields.Order)
	useOffsetParam(q, p.Offset)
	useLimitParam(q, p.Limit)
}

func useFulltextParam(q *esquel.Esquel, fulltext string, columns []string) {
	if len(fulltext) == 0 || len(columns) == 0 {
		return
	}
	startWord := "WHERE"
	if q.WhereExists() {
		startWord = "AND"
	}
	conditions := make([]string, len(columns))
	args := make(esquel.Map)
	for i := range conditions {
		name := fmt.Sprintf("fulltext%d", i+1)
		args[name] = esquel.CreateTsQuery(fulltext)
		conditions[i] = columns[i] + " @@ to_tsquery(@" + name + ")"
	}
	q.Q(startWord+` (`+strings.Join(conditions, " OR ")+`)`, args)
}

func useOrderParam(q *esquel.Esquel, order []string, columns map[string]string) {
	if len(order) == 0 || len(columns) == 0 {
		return
	}
	r := make([]string, 0)
	for _, o := range order {
		if !strings.Contains(o, ":") {
			continue
		}
		parts := strings.Split(o, ":")
		if len(parts) < 2 || parts[1] == "" {
			continue
		}
		column, ok := columns[parts[0]]
		if !ok {
			continue
		}
		r = append(r, strx.Escape(column)+" "+strx.Escape(strings.ToUpper(parts[1])))
	}
	if len(r) == 0 {
		return
	}
	q.Q(`ORDER BY ` + strings.Join(r, ","))
}

func useOffsetParam(q *esquel.Esquel, offset int) {
	q.Q(`OFFSET @offset`, esquel.Map{Offset: offset})
}

func useGroupParam(q *esquel.Esquel, group []string) {
	n := len(group)
	if n == 0 {
		return
	}
	r := make([]string, n)
	for i, g := range group {
		r[i] = strx.Escape(strings.TrimSpace(g))
	}
	q.Q(`GROUP BY ` + strings.Join(r, ","))
}

func useLimitParam(q *esquel.Esquel, limit int) {
	if limit == -1 {
		return
	}
	if limit == 0 {
		limit = DefaultLimit
	}
	q.Q(`LIMIT @limit`, esquel.Map{Limit: limit})
}
