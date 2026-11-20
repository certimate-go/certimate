package installer

import (
	"context"
)

type superuserRepository interface {
	IsEmpty(ctx context.Context) (bool, error)
	Create(ctx context.Context, email, password string) error
}
