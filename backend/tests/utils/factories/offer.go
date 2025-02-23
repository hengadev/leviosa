package factories

import (
	"github.com/GaryHY/leviosa/internal/domain/product"
)

func NewBasicOffer(overrides map[string]interface{}) *productService.Offer {
	offer := &productService.Offer{
		ID:          "f1625792-6363-4111-943a-547d68d76d1t",
		ProductID:   "366eb4a7-7853-4854-ab08-d12c53af7503",
		Name:        "First offer name for Leviosa",
		Description: "First offer description for Leviosa",
		Picture:     "picture",
		Duration:    30,
		Price:       18,
		PriceID:     "07df7a28-f5f0-48ac-a11a-3d8b96d7760e",
	}
	// Apply overrides
	for key, value := range overrides {
		switch key {
		case "ID":
			offer.ID = value.(string)
		case "ProductID":
			offer.ProductID = value.(string)
		case "Name":
			offer.Name = value.(string)
		case "Description":
			offer.Description = value.(string)
		case "Picture":
			offer.Picture = value.(string)
		case "Duration":
			offer.Duration = value.(int)
		case "Price":
			offer.Price = value.(int64)
		case "PriceID":
			offer.PriceID = value.(string)
		}
	}
	return offer
}

func NewBasicOfferList() []*productService.Offer {
	offers := []*productService.Offer{
		NewBasicOffer(nil),
		NewBasicOffer(map[string]any{
			"ID":          "42002e73-cfd4-4e2b-a914-06ddee24823a",
			"ProductID":   "67c64adb-cf07-423d-9500-9211a139dacf",
			"Name":        "Second offer name for Leviosa",
			"Description": "Second offer description for Leviosa",
			"Picture":     "picture2",
			"Duration":    30,
			"Price":       20,
			"PriceID":     "1fe9daa1-b6c9-4c1f-8ec2-a3c8274c6211",
		}),
		NewBasicOffer(map[string]any{
			"ID":          "472f99f7-b15f-4dee-8e49-2fc28c8cd28e",
			"ProductID":   "5983ef96-5b98-4267-ba44-4abb6ad8f6db",
			"Name":        "Third offer name for Leviosa",
			"Description": "Third offer description for Leviosa",
			"Picture":     "picture3",
			"Duration":    10,
			"Price":       50,
			"PriceID":     "f148df20-b6c4-4e9e-9f3d-5b88ac80c4f7",
		}),
	}
	return offers
}
