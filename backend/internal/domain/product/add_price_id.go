package productService

import "context"

func (s *Service) AddPriceID(ctx context.Context, productID, priceID string) error {
	if err := s.repo.AddPriceID(ctx, productID, priceID); err != nil {
		switch {
		}
	}
	return nil
}
