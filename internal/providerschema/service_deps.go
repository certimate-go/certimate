package providerschema

import "context"

type providerSchemaRepository interface {
	Get(ctx context.Context, providerType string) (*Schema, error)
	List(ctx context.Context) ([]*Schema, error)
}
