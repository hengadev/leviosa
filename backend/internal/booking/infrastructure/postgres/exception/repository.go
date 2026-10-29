// Package exceptionRepository stores availability exceptions (issue 012) in
// booking.availability_exceptions.
package exceptionRepository

import (
	"context"
	"fmt"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool   *pgxpool.Pool
	schema string
}

func New(ctx context.Context, pool *pgxpool.Pool) ports.AvailabilityExceptionRepository {
	return &Repository{pool: pool, schema: "booking"}
}

// Dates are written as YYYY-MM-DD text so the stored day never depends on a
// time zone; DATE columns scan back as midnight UTC, the civil date
// representation of the domain.

func (r *Repository) Create(ctx context.Context, exception *domain.AvailabilityException) error {
	query := fmt.Sprintf(`
		INSERT INTO %s.availability_exceptions (id, availability_id, start_date, end_date, created_at)
		VALUES ($1, $2, $3::date, $4::date, $5)
	`, r.schema)
	_, err := r.pool.Exec(ctx, query,
		exception.ID,
		exception.AvailabilityID,
		exception.StartDate.Format(time.DateOnly),
		exception.EndDate.Format(time.DateOnly),
		exception.CreatedAt,
	)
	if err != nil {
		return errs.ClassifyPgError("create availability exception", err)
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*domain.AvailabilityException, error) {
	query := fmt.Sprintf(`
		SELECT id, availability_id, start_date, end_date, created_at
		FROM %s.availability_exceptions
		WHERE id = $1
	`, r.schema)
	exception, err := scanException(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		return nil, errs.ClassifyPgError("get availability exception", err)
	}
	return &exception, nil
}

func (r *Repository) ListByAvailabilityIDs(ctx context.Context, availabilityIDs []uuid.UUID) ([]domain.AvailabilityException, error) {
	if len(availabilityIDs) == 0 {
		return nil, nil
	}
	query := fmt.Sprintf(`
		SELECT id, availability_id, start_date, end_date, created_at
		FROM %s.availability_exceptions
		WHERE availability_id = ANY($1)
		ORDER BY start_date, end_date, created_at
	`, r.schema)
	rows, err := r.pool.Query(ctx, query, availabilityIDs)
	if err != nil {
		return nil, errs.ClassifyPgError("list availability exceptions", err)
	}
	defer rows.Close()

	exceptions := []domain.AvailabilityException{}
	for rows.Next() {
		exception, err := scanException(rows)
		if err != nil {
			return nil, errs.ClassifyPgError("scan availability exception", err)
		}
		exceptions = append(exceptions, exception)
	}
	if err := rows.Err(); err != nil {
		return nil, errs.ClassifyPgError("list availability exceptions", err)
	}
	return exceptions, nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	query := fmt.Sprintf(`DELETE FROM %s.availability_exceptions WHERE id = $1`, r.schema)
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return errs.ClassifyPgError("delete availability exception", err)
	}
	if result.RowsAffected() == 0 {
		return errs.ErrRepositoryNotFound
	}
	return nil
}

func scanException(row pgx.Row) (domain.AvailabilityException, error) {
	var exception domain.AvailabilityException
	err := row.Scan(&exception.ID, &exception.AvailabilityID, &exception.StartDate, &exception.EndDate, &exception.CreatedAt)
	if err != nil {
		return domain.AvailabilityException{}, err
	}
	exception.StartDate = civil(exception.StartDate)
	exception.EndDate = civil(exception.EndDate)
	return exception, nil
}

func civil(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
