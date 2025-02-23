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

func NewBasicProductList() []*productService.Product {
	products := []*productService.Product{
		NewBasicProduct(nil),
		NewBasicProduct(map[string]any{
			"ID":          "893a7ff5-bc34-438a-a0ed-1d426711e77a",
			"Name":        "Second product name for Leviosa",
			"Description": "Second product description for Leviosa",
		}),
		NewBasicProduct(map[string]any{
			"ID":          "019575bc-8494-45d9-9ca2-85bafa86a64f",
			"Name":        "Third product name for Leviosa",
			"Description": "Third product description for Leviosa",
		}),
	}
	return products
}
