package connection

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/Blackmocca/go-lightweight-scheduler/service/v1/schedule"
)

var (
	path_database = "./migrations/database"
	path_table    = func(dbtype constants.AdapterDatabaseConnectionType) string {
		return fmt.Sprintf("%s/%s", path_database, string(dbtype))
	}
)

type DatabaseAdapterConnection interface {
	GetClient() interface{}
	SetClient(ctx context.Context, db interface{})
	GetConnectionURI() string
	GetDatabaseType() constants.AdapterDatabaseConnectionType
	GetRepository() schedule.Repository
	IsConnect(ctx context.Context) bool
	Close(ctx context.Context) error
}

func GetDatabaseConnection(ctx context.Context, adapter constants.AdapterDatabaseConnectionType, uri string) (DatabaseAdapterConnection, error) {
	switch adapter {
	case constants.ADAPTER_DATABASE_POSTGRES:
		client, err := NewPsqlConnection(ctx, uri)
		if err != nil {
			return nil, err
		}
		return client, nil
	case constants.ADAPTER_DATABASE_MONGODB:
		break
	}
	return nil, fmt.Errorf("%s not found", adapter)
}

func MigrateUp(adapter DatabaseAdapterConnection) error {
	path := path_table(adapter.GetDatabaseType())
	return runcmd(adapter.GetConnectionURI(), path, "up")
}

// operation [up|down]
func runcmd(dbcon string, path string, operation string) error {
	execute := fmt.Sprintf(`migrate -database "%s" -path "%s" %s`, dbcon, path, operation)
	fmt.Println(fmt.Sprintf("running command `%s`", execute))

	cmd := exec.Command("bash", "-c", execute)
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}
