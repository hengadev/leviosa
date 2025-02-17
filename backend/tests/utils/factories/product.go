package factories

import (
	"github.com/GaryHY/leviosa/internal/domain/product"
)

func NewBasicProduct(overrides map[string]interface{}) *productService.Product {
	product := &productService.Product{
		ID:          "f1625792-6363-4111-943a-547d68d76d15",
		Name:        "First product name for Leviosa",
		Description: "First product description for Leviosa",
	}
	// Apply overrides
	for key, value := range overrides {
		switch key {
		case "ID":
			product.ID = value.(string)
		case "Name":
			product.Name = value.(string)
		case "Description":
			product.Description = value.(string)
		}
	}
	return product
}
