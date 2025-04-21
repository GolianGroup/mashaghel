package arango

import (
	"context"
	"fmt"
	"mashaghel/internal/config"
	"time"

	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/connection"
)

type ArangoDB interface {
	Database(ctx context.Context) arangodb.Database
	GetCollection(ctx context.Context, name string) (arangodb.Collection, error)
	Ping(ctx context.Context) error
}

type arangoDB struct {
	database arangodb.Database
	client   arangodb.Client
	config   *config.ArangoConfig
}

func NewArangoDB(ctx context.Context, conf *config.ArangoConfig) (ArangoDB, error) {
	connStrs, err := config.GetArangoStrings(conf)
	if err != nil {
		return nil, err
	}
	endpoint := connection.NewRoundRobinEndpoints(connStrs)
	conn := connection.NewHttp2Connection(connection.DefaultHTTP2ConfigurationWrapper(endpoint /*InsecureSkipVerify*/, conf.InsecureSkipVerify))

	auth := connection.NewBasicAuth(conf.User, conf.Pass)
	err = conn.SetAuthentication(auth)
	if err != nil {
		return nil, err
	}

	client := arangodb.NewClient(conn)

	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	dbExists, err := client.DatabaseExists(timeoutCtx, conf.DBName)
	if err != nil {
		return nil, err
	}
	if !dbExists {
		return nil, fmt.Errorf("database %s does not exist", conf.DBName)
	}

	db, err := client.Database(timeoutCtx, conf.DBName)
	if err != nil {
		return nil, err
	}

	return &arangoDB{
		database: db,
		client:   client,
		config:   conf,
	}, nil
}

func (a *arangoDB) Ping(ctx context.Context) error {
	_, err := a.database.Info(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (a *arangoDB) Database(ctx context.Context) arangodb.Database {
	return a.database
}

func (a *arangoDB) GetCollection(ctx context.Context, name string) (arangodb.Collection, error) {
	options := arangodb.GetCollectionOptions{
		SkipExistCheck: false,
	}

	return a.database.GetCollection(ctx, name, &options)
}
