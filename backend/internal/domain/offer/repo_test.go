package offerService_test

import (
	"context"
	"github.com/GaryHY/event-reservation-app/internal/domain/offer"
)

type StubOfferRepository struct {
	offers []*offerService.Offer
}

func (s *StubOfferRepository) GetAllOffers(ctx context.Context) ([]*offerService.Offer, error) {
	return nil, nil
}
func (s *StubOfferRepository) CreateOffer(ctx context.Context, offer *offerService.Offer) error {
	return nil
}
