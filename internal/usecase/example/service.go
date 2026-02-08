package example

import (
	"context"

	"github.com/b92c/go-boilerplate/internal/adapter/repository/dynamo"
	"github.com/b92c/go-boilerplate/internal/port"
)

// Service expõe operações de domínio usando um repositório que depende da porta de DynamoDB.
type Service struct {
	repo *dynamo.ExampleRepository
	log  port.Logger
}

func NewService(repo *dynamo.ExampleRepository, log port.Logger) *Service {
	return &Service{repo: repo, log: log}
}

func (s *Service) Create(ctx context.Context, item map[string]any) error {
	if s.log != nil {
		s.log.Info("create item", "table", s.repo.TableName())
	}
	return s.repo.Create(ctx, item)
}

func (s *Service) Get(ctx context.Context, key map[string]any) (map[string]any, error) {
	return s.repo.Get(ctx, key)
}

func (s *Service) Update(ctx context.Context, item map[string]any) error {
	return s.repo.Update(ctx, item)
}

func (s *Service) Delete(ctx context.Context, key map[string]any) error {
	return s.repo.Delete(ctx, key)
}

func (s *Service) List(ctx context.Context, limit int32) ([]map[string]any, error) {
	return s.repo.List(ctx, limit)
}
