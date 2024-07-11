package convertx

import (
	"database/sql"
	"reflect"
	"strconv"
	"time"
)

var (
	nullStringType  = reflect.TypeOf(sql.Null[string]{})
	nullIntType     = reflect.TypeOf(sql.Null[int]{})
	nullInt64Type   = reflect.TypeOf(sql.Null[int64]{})
	nullFloat32Type = reflect.TypeOf(sql.Null[float32]{})
	nullFloat64Type = reflect.TypeOf(sql.Null[float64]{})
	nullBoolType    = reflect.TypeOf(sql.Null[bool]{})
	nullTimeType    = reflect.TypeOf(sql.Null[time.Time]{})
)

func ConvertSqlNullToString(value any) string {
	switch v := value.(type) {
	case sql.Null[string]:
		return v.V
	case sql.Null[int]:
		return strconv.Itoa(v.V)
	case sql.Null[int64]:
		return strconv.FormatInt(v.V, 10)
	case sql.Null[float32]:
		return strconv.FormatFloat(float64(v.V), 'f', -1, 32)
	case sql.Null[float64]:
		return strconv.FormatFloat(v.V, 'f', -1, 64)
	case sql.Null[bool]:
		return strconv.FormatBool(v.V)
	case sql.Null[time.Time]:
		return v.V.UTC().String()
	default:
		return ""
	}
}

func ConvertStringToSqlNull(value string, valueType reflect.Type) any {
	switch valueType {
	case nullStringType:
		return sql.Null[string]{V: value, Valid: len(value) > 0}
	case nullInt64Type, nullIntType:
		if len(value) == 0 {
			value = "0"
		}
		res, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return sql.Null[any]{V: nil, Valid: false}
		}
		return sql.Null[int64]{V: res, Valid: res > 0}
	case nullFloat32Type:
		if len(value) == 0 {
			value = "0"
		}
		res, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return sql.Null[any]{V: nil, Valid: false}
		}
		return sql.Null[float32]{V: float32(res), Valid: res > 0}
	case nullFloat64Type:
		if len(value) == 0 {
			value = "0"
		}
		res, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return sql.Null[any]{V: nil, Valid: false}
		}
		return sql.Null[float64]{V: res, Valid: res > 0}
	case nullBoolType:
		valid := false
		if value == "true" || value == "false" {
			valid = true
		}
		return sql.Null[bool]{V: value == "true", Valid: valid}
	case nullTimeType:
		parsedTime := ConvertStringToTime(value)
		return sql.Null[time.Time]{V: parsedTime, Valid: !parsedTime.IsZero()}
	default:
		return sql.Null[any]{V: nil, Valid: false}
	}
}
