package offerService

import (
	"context"
)

type Reader interface {
	GetAllOffers(ctx context.Context) ([]*Offer, error)
}

type Writer interface {
	CreateOffer(ctx context.Context, offer *Offer) error
	DeleteOffer(ctx context.Context, name string) error
}

type ReadWriter interface {
	Reader
	Writer
}
