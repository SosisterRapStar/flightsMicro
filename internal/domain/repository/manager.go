package repository

import (
	"context"
)

type Manager interface {
	Do(context.Context, func(context.Context) error) error
}
