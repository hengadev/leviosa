package productService

import (
	"context"

	"github.com/GaryHY/leviosa/pkg/errsx"
)

// (massage, coaching mental etc...)
type ProductType struct {
	ID          int
	Name        string
	Description string
}

func (p ProductType) Valid(ctx context.Context) (problems errsx.Map) {
	var errs errsx.Map
	return errs
}

func (p ProductType) AssertComparable() {}

func (p ProductType) GetSQLColumnMapping() map[string]string {
	return map[string]string{
		"ID":          "id",
		"Name":        "name",
		"Description": "description",
	}
}

func (p ProductType) GetProhibitedFields() []string {
	return []string{
		"ID",
	}
}
