package dynamo

import (
	"context"

	"github.com/b92c/go-boilerplate/internal/port"
)

type ExampleRepository struct {
	db    port.DynamoDBPort
	table string
}

func NewExampleRepository(db port.DynamoDBPort, table string) *ExampleRepository {
	return &ExampleRepository{db: db, table: table}
}

func (r *ExampleRepository) TableName() string { return r.table }

func (r *ExampleRepository) Create(ctx context.Context, item map[string]any) error {
	return r.db.PutItem(ctx, r.table, item)
}

func (r *ExampleRepository) Get(ctx context.Context, key map[string]any) (map[string]any, error) {
	return r.db.GetItem(ctx, r.table, key)
}

func (r *ExampleRepository) Update(ctx context.Context, item map[string]any) error {
	return r.db.PutItem(ctx, r.table, item)
}

func (r *ExampleRepository) Delete(ctx context.Context, key map[string]any) error {
	return r.db.DeleteItem(ctx, r.table, key)
}

func (r *ExampleRepository) List(ctx context.Context, limit int32) ([]map[string]any, error) {
	return r.db.Scan(ctx, r.table, limit)
}
