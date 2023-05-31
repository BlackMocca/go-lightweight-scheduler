package connection

import (
	"context"
	"fmt"

	"github.com/BlackMocca/sqlx"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/Blackmocca/go-lightweight-scheduler/service/v1/schedule"
	"github.com/Blackmocca/go-lightweight-scheduler/service/v1/schedule/repository"
	pg "github.com/lib/pq"
)

const (
	postgres_driver = "postgres"
)

type PGClient struct {
	db            *sqlx.DB
	connectionURI string
	driverName    string
	repository    schedule.Repository
}

func NewPsqlConnection(ctx context.Context, connectionStr string) (*PGClient, error) {
	addr, err := pg.ParseURL(connectionStr)
	if err != nil {
		return nil, err
	}

	db, err := sqlx.Connect(postgres_driver, addr)
	if err != nil {
		return nil, err
	}

	client := &PGClient{
		db:            db,
		connectionURI: connectionStr,
		driverName:    postgres_driver,
	}

	if !client.IsConnect(ctx) {
		return nil, fmt.Errorf("can't connect database with connection string: %s", connectionStr)
	}

	client.repository = repository.NewPsqlRepository(client.db)
	return client, nil
}

func (c *PGClient) GetClient() interface{} {
	return c.db
}

func (c *PGClient) GetConnectionURI() string {
	return c.connectionURI
}

func (c *PGClient) GetDatabaseType() constants.AdapterDatabaseConnectionType {
	return constants.ADAPTER_DATABASE_POSTGRES
}

func (c *PGClient) GetRepository() schedule.Repository {
	return c.repository
}

func (c *PGClient) SetClient(ctx context.Context, db interface{}) {
	c.db = db.(*sqlx.DB)
}

func (c *PGClient) IsConnect(ctx context.Context) bool {
	if err := c.db.Ping(); err == nil {
		return true
	}
	return false
}

func (c *PGClient) Close(ctx context.Context) error {
	return c.db.Close()
}
