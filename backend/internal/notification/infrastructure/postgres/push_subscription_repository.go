// Package postgres stores the notification domain's records: the Web Push
// subscriptions of practitioner notifications (ADR-0020), in
// notification.push_subscriptions, encrypted with encx.
package postgres

import (
	"context"

	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/notification/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/notification/ports"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PushSubscriptionRepository is the Postgres push subscription store.
type PushSubscriptionRepository struct {
	pool *pgxpool.Pool
}

var _ ports.PushSubscriptionRepository = (*PushSubscriptionRepository)(nil)

func NewPushSubscriptionRepository(pool *pgxpool.Pool) *PushSubscriptionRepository {
	return &PushSubscriptionRepository{pool: pool}
}

// Upsert inserts the subscription, or takes over the row of the same device
// (same endpoint hash): a device belongs to the last user who switched
// notifications on with it.
func (r *PushSubscriptionRepository) Upsert(ctx context.Context, sub *domain.PushSubscriptionEncx) error {
	const query = `
		INSERT INTO notification.push_subscriptions (
			id, user_id, endpoint_hash, endpoint_encrypted, p256dh_encrypted, auth_encrypted,
			user_agent, dek_encrypted, key_version, metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (endpoint_hash) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			endpoint_encrypted = EXCLUDED.endpoint_encrypted,
			p256dh_encrypted = EXCLUDED.p256dh_encrypted,
			auth_encrypted = EXCLUDED.auth_encrypted,
			user_agent = EXCLUDED.user_agent,
			dek_encrypted = EXCLUDED.dek_encrypted,
			key_version = EXCLUDED.key_version,
			metadata = EXCLUDED.metadata
	`
	_, err := r.pool.Exec(ctx, query,
		sub.ID,
		sub.UserID,
		sub.EndpointHash,
		sub.EndpointEncrypted,
		sub.P256dhEncrypted,
		sub.AuthEncrypted,
		sub.UserAgent,
		sub.DEKEncrypted,
		sub.KeyVersion,
		sub.Metadata,
		sub.CreatedAt,
	)
	if err != nil {
		return errs.ClassifyPgError("upsert push subscription", err)
	}
	return nil
}

func (r *PushSubscriptionRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.PushSubscriptionEncx, error) {
	const query = `
		SELECT id, user_id, endpoint_hash, endpoint_encrypted, p256dh_encrypted, auth_encrypted,
			user_agent, dek_encrypted, key_version, created_at
		FROM notification.push_subscriptions
		WHERE user_id = $1
		ORDER BY created_at, id
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, errs.ClassifyPgError("list push subscriptions", err)
	}
	defer rows.Close()

	subs := []*domain.PushSubscriptionEncx{}
	for rows.Next() {
		sub := &domain.PushSubscriptionEncx{}
		if err := rows.Scan(
			&sub.ID,
			&sub.UserID,
			&sub.EndpointHash,
			&sub.EndpointEncrypted,
			&sub.P256dhEncrypted,
			&sub.AuthEncrypted,
			&sub.UserAgent,
			&sub.DEKEncrypted,
			&sub.KeyVersion,
			&sub.CreatedAt,
		); err != nil {
			return nil, errs.ClassifyPgError("scan push subscription", err)
		}
		subs = append(subs, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, errs.ClassifyPgError("list push subscriptions", err)
	}
	return subs, nil
}

func (r *PushSubscriptionRepository) DeleteByEndpointHash(ctx context.Context, userID uuid.UUID, endpointHash string) error {
	const query = `DELETE FROM notification.push_subscriptions WHERE user_id = $1 AND endpoint_hash = $2`
	if _, err := r.pool.Exec(ctx, query, userID, endpointHash); err != nil {
		return errs.ClassifyPgError("delete push subscription", err)
	}
	return nil
}

func (r *PushSubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `DELETE FROM notification.push_subscriptions WHERE id = $1`
	if _, err := r.pool.Exec(ctx, query, id); err != nil {
		return errs.ClassifyPgError("delete push subscription", err)
	}
	return nil
}
