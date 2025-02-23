package sqliteutil

import (
	"fmt"
	"reflect"
	"strings"
)

// Return the necessary elements to write a query update to target the non zero value.
func WriteInsertQuery[S any](s S) (string, []any) {
	var tables []string
	var values []any
	v := reflect.ValueOf(s)
	t := reflect.TypeOf(s)
	vf := reflect.VisibleFields(t)
	tableName := fmt.Sprintf("%ss", strings.ToLower(t.Name()))
	query := fmt.Sprintf("UPDATE %s set ", tableName)
	for _, f := range vf {
		value := v.FieldByName(f.Name)
		if !value.IsZero() && value.CanInterface() {
			tables = append(tables, fmt.Sprintf("%s=?", strings.ToLower(f.Name)))
			values = append(values, value.Interface())
		}
	}
	query += strings.Join(tables, ", ")
	query += " WHERE id=?;"
	return query, values
}
