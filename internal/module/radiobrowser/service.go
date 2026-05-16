package radiobrowser

import "context"

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

// ListStationsInput holds pagination data for listing radio browser stations.
type ListStationsInput struct {
	Name     string
	Country  string
	Language string
	Limit    int64
	Offset   int64
}

// Service defines the radio browser station operations.
type Service interface {
	List(ctx context.Context, input ListStationsInput) ([]Station, int64, error)
}

type service struct {
	repo Repository
}

// NewService constructs a Service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, input ListStationsInput) ([]Station, int64, error) {
	total, err := s.repo.Count(ctx, input.Name, input.Country, input.Language)
	if err != nil {
		return nil, 0, err
	}
	rows, err := s.repo.List(ctx, input.Name, input.Country, input.Language, input.Limit, input.Offset)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
