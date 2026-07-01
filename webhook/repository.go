// Use el mismo postegerSQL para almacenar el webhook, no se si tenias otra idea yo lo hice asi
// Recuerda migrar de nuevo si lo vas a mantener x esta via
// La otra via q me dieron es si todo esta almacenado en un mismo webhook hacerlo como env...pero como tu nivel de detalles es corto lo deje para que hubiesen varias webhooks...no se tu idea cual era

// Entonces ahi cree la tablita muy bonita y demas
// Solo hice registrado y un edit rarisimo que me copie de google y delete

package webhook

import (
	"context"
	"database/sql"
)

type WebhookRepository struct {
    db *sql.DB
}

func NewWebhookRepository(db *sql.DB) *WebhookRepository {
	return &WebhookRepository{db: db}
}

func (r *WebhookRepository) CreateTable(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS webhooks (
		id         SERIAL PRIMARY KEY,
		phone      VARCHAR(32) NOT NULL UNIQUE,
		url        TEXT NOT NULL,
		secret     VARCHAR(128) DEFAULT '',
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	)`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *WebhookRepository) Register(ctx context.Context, phone, url, secret string) error {
	query := `
	INSERT INTO webhooks (phone, url, secret)
	VALUES ($1, $2, $3)
	ON CONFLICT (phone) DO UPDATE SET url = $2, secret = $3, updated_at = NOW()`
	_, err := r.db.ExecContext(ctx, query, phone, url, secret)
	return err
}

func (r *WebhookRepository) GetByPhone(ctx context.Context, phone string) (*WebhookEntry, error) {
	query := `SELECT id, phone, url, secret, created_at FROM webhooks WHERE phone = $1`
	row := r.db.QueryRowContext(ctx, query, phone)

	var e WebhookEntry
	err := row.Scan(&e.ID, &e.Phone, &e.URL, &e.Secret, &e.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (r *WebhookRepository) Delete(ctx context.Context, phone string) error {
	query := `DELETE FROM webhooks WHERE phone = $1`
	_, err := r.db.ExecContext(ctx, query, phone)
	return err
}

