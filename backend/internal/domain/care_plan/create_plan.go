package carePlanService

import "context"

func (s *Service) CreatePlan(
	ctx context.Context,
	registrationID string,
	feedback string,
	exercices []string,
	videos []string,
) error {
	return nil
}
