package connection

import (
	"context"
	"fmt"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
)

type DatabaseAdapterConnection interface {
	GetClient() interface{}
	SetClient(ctx context.Context, db interface{})
	GetConnectionURI() string
	IsConnect(ctx context.Context) bool
	Close(ctx context.Context) error
}

func GetDatabaseConnection(ctx context.Context, adapter constants.AdapterDatabaseConnectionType, uri string) (DatabaseAdapterConnection, error) {
	switch adapter {
	case constants.ADAPTER_DATABASE_POSTGRES:
		return NewPsqlConnection(ctx, uri)
	case constants.ADAPTER_DATABASE_MONGODB:
		break
	}
	return nil, fmt.Errorf("%s not found", adapter)
}
