package dyna

import (
	"fmt"
	"strings"
	
	"github.com/daarlabs/hirokit/esquel"
	"github.com/daarlabs/hirokit/hiro"
	"github.com/daarlabs/hirokit/util/strx"
)

type Param struct {
	Fulltext  string
	Offset    int
	Limit     int
	Order     []string
	Group     []string
	Filter    []string
	FilterMap hiro.Map
	Fields    Fields
}

type Fields struct {
	Fulltext []string
	Map      map[string]string
}

func (p Param) Parse(c hiro.Ctx) Param {
	if p.FilterMap == nil {
		p.FilterMap = make(hiro.Map)
	}
	c.Parse().MustQuery(Fulltext, &p.Fulltext)
	c.Parse().MustQuery(Offset, &p.Offset)
	c.Parse().MustQuery(Limit, &p.Limit)
	c.Parse().Multiple().MustQuery(Order, &p.Order)
	for filterKey := range p.Fields.Map {
		filterValue := make([]string, 0)
		c.Parse().Multiple().MustQuery(filterKey, &filterValue)
		for _, v := range filterValue {
			if len(v) == 0 {
				continue
			}
			p.Filter = append(p.Filter, fmt.Sprintf("%s:%s", filterKey, v))
			p.FilterMap[filterKey] = v
		}
	}
	return p
}

func (p Param) Use(q *esquel.Esquel) {
	useFulltextParam(q, p.Fields.Map, p.Fulltext, p.Fields.Fulltext)
	useFilterParam(q, p.Filter, p.Fields.Map)
	useGroupParam(q, p.Group)
	useOrderParam(q, p.Order, p.Fields.Map)
	useOffsetParam(q, p.Offset)
	useLimitParam(q, p.Limit)
}

func useFulltextParam(q *esquel.Esquel, fields map[string]string, fulltext string, columns []string) {
	if len(fulltext) == 0 || len(columns) == 0 {
		return
	}
	hasWildcard := strings.Contains(fulltext, ":*")
	isLike := strings.Contains(fulltext, ":") && !hasWildcard
	startWord := "WHERE"
	if q.WhereExists() {
		startWord = "AND"
	}
	var n int
	if !isLike {
		n = len(columns)
	}
	if isLike {
		n = len(strings.Split(fulltext, ";"))
	}
	conditions := make([]string, n)
	args := make(esquel.Map)
	if !isLike {
		for i := range conditions {
			name := fmt.Sprintf("fulltext%d", i+1)
			args[name] = esquel.CreateTsQuery(fulltext)
			conditions[i] = columns[i] + " @@ to_tsquery(@" + name + ")"
		}
	}
	if isLike {
		for i, item := range strings.Split(fulltext, ";") {
			parts := strings.Split(item, ":")
			if len(parts) < 2 {
				continue
			}
			field := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			args["like-"+field] = value
			f := field
			if v, ok := fields[field]; ok {
				f = v
			}
			conditions[i] = `lower(unaccent(` + f + `))` + ` LIKE CONCAT('%',lower(unaccent(@like-` + field + `)),'%')`
		}
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
		field, ok := columns[parts[0]]
		if !ok {
			continue
		}
		r = append(r, strx.Escape(field)+" "+strx.Escape(strings.ToUpper(parts[1])))
	}
	if len(r) == 0 {
		return
	}
	q.Q(`ORDER BY ` + strings.Join(r, ","))
}

func useFilterParam(q *esquel.Esquel, filter []string, columns map[string]string) {
	if len(filter) == 0 || len(columns) == 0 {
		return
	}
	conditions := make([]string, 0)
	args := make(esquel.Map)
	for _, f := range filter {
		if !strings.Contains(f, ":") {
			continue
		}
		parts := strings.Split(f, ":")
		if len(parts) < 2 || parts[1] == "" {
			continue
		}
		field, ok := columns[parts[0]]
		if !ok {
			continue
		}
		name := f[:strings.Index(f, ":")]
		switch parts[1] {
		case "0":
			continue
		case "on":
			args["filter-"+name] = true
		default:
			args["filter-"+name] = parts[1]
		}
		conditions = append(conditions, field+" = @filter-"+name)
	}
	if len(conditions) == 0 {
		return
	}
	startWord := "WHERE"
	if q.WhereExists() {
		startWord = "AND"
	}
	q.Q(startWord+` (`+strings.Join(conditions, " AND ")+`)`, args)
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
