package carePlanService

import (
	"context"
)

type Reader interface{}

type Writer interface {
	CreatePlan(ctx context.Context, registrationID string, content string) error
}

type ReadWriter interface {
	Reader
	Writer
}
