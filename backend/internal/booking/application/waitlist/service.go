package waitlist

import (
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	catalogPorts "github.com/Leviosa-care/leviosa/backend/internal/catalog/ports"
	"github.com/hengadev/encx"
)

// WaitlistService implements ports.WaitlistService.
type WaitlistService struct {
	waitlistRepo   ports.WaitlistRepository
	productService catalogPorts.PublicProductService
	crypto         encx.CryptoService
}

// New creates a new instance of the waitlist service.
func New(waitlistRepo ports.WaitlistRepository, productService catalogPorts.PublicProductService, crypto encx.CryptoService) ports.WaitlistService {
	return &WaitlistService{
		waitlistRepo:   waitlistRepo,
		productService: productService,
		crypto:         crypto,
	}
}
