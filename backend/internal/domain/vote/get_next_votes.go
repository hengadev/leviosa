package vote

import (
	"context"
	"fmt"
	"time"
)

func (s *Service) GetNextVotes(ctx context.Context) ([]*Vote, error) {
	now := time.Now().UTC()
	day := int(now.Day())
	month := int(now.Month())
	year := int(now.Year())
	votes, err := s.Repo.GetNextVotes(ctx, day, month, year)
	if err != nil {
		return nil, fmt.Errorf("find next votes: %w", err)
	}
	return votes, nil
}
