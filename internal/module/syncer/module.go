package syncer

import (
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/radiobrowser"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Module wires syncer dependencies together.
type Module struct {
	svc Service
}

// NewModule constructs a syncer Module.
func NewModule(pool *pgxpool.Pool, client *radiobrowser.Client) *Module {
	repo := NewRepository(pool)
	svc := newService(repo, client)
	return &Module{svc: svc}
}

// Service returns the Service for use by the cron entrypoint.
func (m *Module) Service() Service {
	return m.svc
}
