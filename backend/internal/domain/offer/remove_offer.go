package offerService

import "context"

// func (s *Service) RemoveOffer(ctx context.Context, offer *Offer) error {
func (s *Service) RemoveOffer(ctx context.Context, name string) error {
	// TODO: remove the offer from the database and return an error if that name is not found in the database
	if err := s.repo.DeleteOffer(ctx, name); err != nil {
		// TODO: make a better error message
		return err
	}
	return nil
}
