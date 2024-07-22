package dyna

import (
	"cmp"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	
	"github.com/daarlabs/hirokit/esquel"
	"github.com/daarlabs/hirokit/util/strx"
)

type Dyna interface {
	DB(db *esquel.DB, query Query) Dyna
	Data(data []map[string]any) Dyna
	
	FindFunc(fn func(param Param, t any) error) Dyna
	FindOneFunc(fn func(name string, v any, t any) error) Dyna
	FindManyFunc(fn func(name string, v any, t any) error) Dyna
	
	Find(param Param, t any) error
	FindOne(name string, v any, t any) error
	FindMany(name string, v any, t any) error
	
	MustFind(param Param, t any)
	MustFindOne(name string, v any, t any)
	MustFindMany(name string, v any, t any)
}

type dyna struct {
	db           *esquel.DB
	data         []map[string]any
	query        Query
	findFunc     func(param Param, t any) error
	findOneFunc  func(name string, v any, t any) error
	findManyFunc func(name string, v any, t any) error
}

func New() Dyna {
	return &dyna{
		data: make([]map[string]any, 0),
	}
}

func (d *dyna) DB(db *esquel.DB, query Query) Dyna {
	d.db = db
	d.query = query
	return d
}

func (d *dyna) Data(data []map[string]any) Dyna {
	d.data = data
	return d
}

func (d *dyna) FindFunc(fn func(param Param, t any) error) Dyna {
	d.findFunc = fn
	return d
}

func (d *dyna) FindOneFunc(fn func(name string, v any, t any) error) Dyna {
	d.findOneFunc = fn
	return d
}

func (d *dyna) FindManyFunc(fn func(name string, v any, t any) error) Dyna {
	d.findManyFunc = fn
	return d
}

func (d *dyna) Find(param Param, t any) error {
	if param.Limit == 0 {
		param.Limit = DefaultLimit
	}
	if d.findFunc != nil {
		return d.findFunc(param, t)
	}
	if d.shouldUseDb() {
		return d.findWithDb(param, t)
	}
	return d.findWithData(param, t)
}

func (d *dyna) FindOne(name string, v any, t any) error {
	if d.findOneFunc != nil {
		return d.findOneFunc(name, v, t)
	}
	if d.shouldUseDb() {
		return d.findOneWithDb(name, v, t)
	}
	return d.findOneWithData(name, v, t)
}

func (d *dyna) FindMany(name string, v any, t any) error {
	if d.findManyFunc != nil {
		return d.findManyFunc(name, v, t)
	}
	if d.shouldUseDb() {
		return d.findManyWithDb(name, v, t)
	}
	return d.findManyWithData(name, v, t)
}

func (d *dyna) MustFind(param Param, t any) {
	if err := d.Find(param, t); err != nil {
		panic(err)
	}
}

func (d *dyna) MustFindOne(name string, v any, t any) {
	if err := d.FindOne(name, v, t); err != nil {
		panic(err)
	}
}

func (d *dyna) MustFindMany(name string, v any, t any) {
	if err := d.FindMany(name, v, t); err != nil {
		panic(err)
	}
}

func (d *dyna) findWithDb(param Param, t any) error {
	q := d.db.Q("SELECT " + d.createFields()).
		Q(`FROM ` + d.query.Table).
		Q(`AS ` + d.query.Alias)
	param.Use(q)
	return q.Exec(t)
}

func (d *dyna) findWithData(param Param, t any) error {
	result := make([]map[string]any, 0)
	for i, row := range d.data {
		if param.Offset != 0 && i < param.Offset {
			continue
		}
		if param.Fulltext != "" {
			shouldContinue := false
			for _, v := range row {
				if strings.Contains(esquel.Normalize(fmt.Sprintf("%v", v)), esquel.Normalize(param.Fulltext)) {
					shouldContinue = true
				}
			}
			if !shouldContinue {
				continue
			}
		}
		result = append(result, row)
		if param.Limit != 0 && i >= param.Limit+param.Offset-1 {
			break
		}
	}
	d.sortDataResult(param, result)
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(resultBytes, t); err != nil {
		return err
	}
	return nil
}

func (d *dyna) sortDataResult(param Param, data []map[string]any) {
	slices.SortFunc(
		data, func(a, b map[string]any) int {
			if len(param.Order) == 0 {
				return 0
			}
			order := param.Order[0]
			parts := strings.Split(order, ":")
			if len(parts) < 2 {
				return 0
			}
			name := strx.Escape(parts[0])
			direction := strx.Escape(parts[1])
			va, oka := a[name]
			vb, okb := b[name]
			if !oka || !okb {
				return 0
			}
			if direction == Asc {
				return cmp.Compare(esquel.Normalize(fmt.Sprintf("%v", va)), esquel.Normalize(fmt.Sprintf("%v", vb)))
			}
			if direction == Desc {
				return cmp.Compare(esquel.Normalize(fmt.Sprintf("%v", vb)), esquel.Normalize(fmt.Sprintf("%v", va)))
			}
			return 0
		},
	)
}

func (d *dyna) findOneWithDb(name string, v any, t any) error {
	fields := make([]string, 0)
	for alias, key := range d.query.Fields {
		fields = append(fields, key+" AS "+alias)
	}
	q := d.db.Q("SELECT "+d.createFields()).
		Q(fmt.Sprintf("FROM %s AS %s", d.query.Table, d.query.Alias)).
		Q(fmt.Sprintf(`WHERE %[1]s = @%[1]s`, name), esquel.Map{name: v}).
		Q(`LIMIT 1`)
	return q.Exec(t)
}

func (d *dyna) findOneWithData(name string, v any, t any) error {
	for _, row := range d.data {
		rv, ok := row[name]
		if !ok {
			continue
		}
		if rv != v {
			continue
		}
		rowBytes, err := json.Marshal(row)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(rowBytes, t); err != nil {
			return err
		}
		break
	}
	return nil
}

func (d *dyna) findManyWithDb(name string, v any, t any) error {
	vv := reflect.ValueOf(v)
	if vv.Kind() != reflect.Slice {
		return ErrorSliceValue
	}
	q := d.db.Q("SELECT "+d.createFields()).
		Q(`FROM `+d.query.Table).
		Q(`WHERE `+name+` IN (@values)`, esquel.Map{"values": v}).
		Q(fmt.Sprintf("LIMIT %d", vv.Len()))
	return q.Exec(t)
}

func (d *dyna) findManyWithData(name string, v any, t any) error {
	vv := reflect.ValueOf(v)
	if vv.Kind() != reflect.Slice {
		return ErrorSliceValue
	}
	result := make([]map[string]any, 0)
	for _, row := range d.data {
		rv, ok := row[name]
		if !ok {
			continue
		}
		var exist bool
		for i := 0; i < vv.Len(); i++ {
			if vv.Index(i).Interface() == rv {
				exist = true
			}
		}
		if !exist {
			continue
		}
		result = append(result, row)
	}
	rowBytes, err := json.Marshal(result)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(rowBytes, t); err != nil {
		return err
	}
	return nil
}

func (d *dyna) shouldUseDb() bool {
	return d.db != nil && d.query.Fields != nil && d.query.Table != ""
}

func (d *dyna) createFields() string {
	fields := make([]string, 0)
	for alias, key := range d.query.Fields {
		fields = append(fields, key+" AS "+alias)
	}
	return strings.Join(fields, ", ")
}
