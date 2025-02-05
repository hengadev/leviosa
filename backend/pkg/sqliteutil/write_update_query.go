package sqliteutil

import (
	"fmt"
	"reflect"
	"strings"
)

type SQLMappable interface {
	GetSQLColumnMapping() map[string]string
}

// Return the necessary elements to write a query update to target the non zero value.
func WriteUpdateQuery[T SQLMappable](
	object T,
	whereMap map[string]any,
	prohibitedFields ...string,
) (string, []any, error) {
	fail := func(err error) (string, []any, error) {
		return "", nil, fmt.Errorf("write update query, %w", err)
	}
	var tables []string
	var values []any
	v := reflect.ValueOf(object)
	t := reflect.TypeOf(object)
	vf := reflect.VisibleFields(t)
	tableName := fmt.Sprintf("%ss", strings.ToLower(t.Name()))
	query := fmt.Sprintf("UPDATE %s set ", tableName)
	for _, f := range vf {
		value := v.FieldByName(f.Name)
		if !value.IsZero() && value.CanInterface() {
			if err := isProhibitedField(f.Name, prohibitedFields...); err != nil {
				return fail(err)
			}
			column := object.GetSQLColumnMapping()[f.Name]
			tables = append(tables, placeholder(column))
			values = append(values, value.Interface())
		}
	}
	query += strings.Join(tables, ", ")
	query += " WHERE "
	var wherePlaceholder []string
	for key, value := range whereMap {
		minKey := strings.ToLower(key)
		wherePlaceholder = append(wherePlaceholder, placeholder(minKey))
		values = append(values, value)
	}
	query += strings.Join(wherePlaceholder, " AND ") + ";"
	return query, values, nil
}

func placeholder(name string) string {
	return fmt.Sprintf("%s=?", name)
}

// Helper function to check if a struct field belongs to a list of strings provided.
func isProhibitedField(name string, prohibitedFields ...string) error {
	for _, prohibitedField := range prohibitedFields {
		if name == prohibitedField {
			return fmt.Errorf("field %q is prohibited", prohibitedField)
		}
	}
	return nil
}
