package sqliteutil

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func BuildDSN(connStr string) string {
	return fmt.Sprintf("%s.db", connStr)
}

func Connect(ctx context.Context, connStr string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}
	return db, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Helper function to check if a struct field belongs to a list of strings provided.
func isProhibitedField(name string, prohibitedFields ...string) error {
	for _, prohibitedField := range prohibitedFields {
		if name == prohibitedField {
			return fmt.Errorf("field %s prohibited", prohibitedField)
		}
	}
	return nil
}

// Return the necessary elements to write a query update to target the non zero value.
func WriteUpdateQuery[T any](
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
			tables = append(tables, placeholder(strings.ToLower(f.Name)))
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
