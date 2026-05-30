package syncer

// Module wires syncer dependencies together.
type Module struct {
	svc Service
}

// NewModule constructs a syncer Module.
func NewModule(svc Service) *Module {
	return &Module{svc: svc}
}

// Service returns the Service for use by the worker entrypoint.
func (m *Module) Service() Service {
	return m.svc
}
