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
