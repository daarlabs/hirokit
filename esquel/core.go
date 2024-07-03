package esquel

import (
	"context"
	"database/sql"
	"reflect"
	"regexp"
	"strings"
	"time"
	
	"github.com/iancoleman/strcase"
	pg "github.com/lib/pq"
)

type Esquel struct {
	*DB
	driverName    string
	dbname        string
	parts         []queryPart
	rows          *sql.Rows
	subscriptions []subscription
}

type Safe []byte

type Map map[string]any

type queryPart struct {
	query string
	arg   Map
}

const (
	querySuffix = ";"
)

var (
	whereFinder = regexp.MustCompile(`\bwhere\b`)
)

func New(db *DB) *Esquel {
	q := &Esquel{
		DB:            db,
		driverName:    db.driverName,
		subscriptions: make([]subscription, 0),
	}
	return q
}

func (q *Esquel) Q(query string, arg ...Map) *Esquel {
	qa := make(Map)
	if len(arg) > 0 {
		qa = arg[0]
	}
	q.parts = append(q.parts, queryPart{query, qa})
	return q
}

func (q *Esquel) WhereExists() bool {
	n := len(q.parts)
	if n > 0 {
		if whereFinder.MatchString(strings.ToLower(q.parts[n-1].query)) {
			return true
		}
	}
	if n > 1 {
		if whereFinder.MatchString(strings.ToLower(q.parts[n-2].query)) {
			return true
		}
	}
	// for _, p := range q.parts {
	// 	if whereFinder.MatchString(strings.ToLower(p.query)) {
	// 		return true
	// 	}
	// }
	return false
}

func (q *Esquel) If(condition bool, query string, arg ...Map) *Esquel {
	if !condition {
		return q
	}
	q.Q(query, arg...)
	return q
}

func (q *Esquel) Subscribe(s subscription) {
	q.subscriptions = append(q.subscriptions, s)
}

func (q *Esquel) CreateSql() string {
	mergedQueryParts, _, err := processQueryParts(q)
	if !strings.HasSuffix(mergedQueryParts, querySuffix) {
		mergedQueryParts += querySuffix
	}
	if err != nil {
		return ""
	}
	return mergedQueryParts
}

func (q *Esquel) CreateMatcher() string {
	return regexp.QuoteMeta(q.CreateSql())
}

func (q *Esquel) Exec(r ...any) error {
	return q.exec(q.DB.context, r...)
}

func (q *Esquel) MustExec(r ...any) {
	if err := q.exec(q.DB.context, r...); err != nil {
		panic(err)
	}
}

func (q *Esquel) ExecCtx(ctx context.Context, r ...any) error {
	return q.exec(ctx, r...)
}

func (q *Esquel) MustExecCtx(ctx context.Context, r ...any) {
	if err := q.exec(ctx, r...); err != nil {
		panic(err)
	}
}

func (q *Esquel) exec(ctx context.Context, result ...any) error {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, q.DB.timeout)
	defer cancel()
	t := time.Now()
	mergedQueryParts, args, err := processQueryParts(q)
	if err != nil {
		return err
	}
	if !strings.HasSuffix(mergedQueryParts, querySuffix) {
		mergedQueryParts += querySuffix
	}
	rows, err := q.DB.QueryContext(ctx, mergedQueryParts, args...)
	if err != nil {
		q.afterQuery(t, mergedQueryParts, args)
		return err
	}
	defer func() {
		_ = rows.Close()
	}()
	if len(result) == 0 {
		q.afterQuery(t, mergedQueryParts, args)
		return nil
	}
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	if len(columns) == 0 {
		q.afterQuery(t, mergedQueryParts, args)
		return nil
	}
	if len(result) > 1 {
		q.scanMultiple(rows, result...)
	}
	if len(result) == 1 {
		q.scanSingle(rows, columns, result[0])
	}
	q.afterQuery(t, mergedQueryParts, args)
	return nil
}

func (q *Esquel) afterQuery(t time.Time, query string, args []any) {
	queryLog := createQueryLog(q.driverName, query, args...)
	duration := time.Now().Sub(t)
	for _, sub := range q.subscriptions {
		sub(queryLog, duration)
	}
	log(q.log, queryLog, duration)
}

func (q *Esquel) scanSingle(rows *sql.Rows, columns []string, result any) {
	res := reflect.ValueOf(result)
	rv := reflect.ValueOf(result)
	rt := reflect.TypeOf(result)
	resKind := rv.Elem().Type().Kind()
	rvKind := rv.Elem().Type().Kind()
	elem := rt.Elem()
	if rvKind == reflect.Slice {
		elem = elem.Elem()
		rv = reflect.New(rv.Elem().Type().Elem())
		rvKind = rv.Elem().Type().Kind()
	}
	for rows.Next() {
		rowData := make([]any, len(columns))
		switch rvKind {
		case reflect.Map:
			rv.Elem().Set(reflect.MakeMap(rv.Elem().Type()))
			for i := range columns {
				field := reflect.New(rv.Elem().Type().Elem())
				rowData[i] = field.Interface()
			}
		case reflect.Struct:
			visibleFields := reflect.VisibleFields(rv.Elem().Type())
			model := make(map[string]any)
			for i := 0; i < rv.Elem().NumField(); i++ {
				field := rv.Elem().Field(i)
				fieldName := rv.Elem().Type().Field(i).Name
				fieldDbName := elem.Field(i).Tag.Get("db")
				exported := false
				for _, vf := range visibleFields {
					if vf.Name == fieldName && vf.IsExported() {
						exported = true
					}
				}
				if !exported {
					continue
				}
				fieldDbNameExists := len(fieldDbName) > 0
				if !fieldDbNameExists {
					fieldName = strcase.ToSnake(fieldName)
				}
				if fieldDbNameExists {
					fieldName = fieldDbName
				}
				switch field.Type().Kind() {
				case reflect.Slice:
					model[fieldName] = pg.Array(field.Addr().Interface())
				default:
					model[fieldName] = field.Addr().Interface()
				}
			}
			for i, c := range columns {
				modelField, ok := model[c]
				if !ok {
					rowData[i] = new(any)
					continue
				}
				rowData[i] = modelField
			}
		default:
			switch rv.Type().Kind() {
			case reflect.Slice:
				rowData[0] = pg.Array(rv.Interface())
			default:
				rowData[0] = rv.Interface()
			}
		}
		if len(rowData) > 0 {
			if scanErr := rows.Scan(rowData...); scanErr != nil {
				panic(scanErr)
			}
		}
		switch resKind {
		case reflect.Slice:
			res.Elem().Set(reflect.Append(res.Elem(), rv.Elem()))
		case reflect.Map:
			for i, c := range columns {
				rv.Elem().SetMapIndex(reflect.ValueOf(c), reflect.ValueOf(rowData[i]).Elem())
			}
		}
	}
}

func (q *Esquel) scanMultiple(rows *sql.Rows, result ...any) {
	resultValues := make([]reflect.Value, len(result))
	for i, r := range result {
		resultValues[i] = reflect.ValueOf(r)
	}
	for rows.Next() {
		rowData := make([]any, 0)
		for _, rv := range resultValues {
			rowData = append(rowData, rv.Interface())
		}
		if len(rowData) > 0 {
			if scanErr := rows.Scan(rowData...); scanErr != nil {
				panic(scanErr)
			}
		}
	}
}
