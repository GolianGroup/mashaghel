package arango

import (
	"context"
	"fmt"
	"log"
	"mashaghel/internal/config"
	"testing"
	"time"

	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/connection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestArangoDB(t *testing.T) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "arangodb:latest",
		ExposedPorts: []string{"8529/tcp"},
		Env: map[string]string{
			"ARANGO_ROOT_PASSWORD": "rootpassword",
		},
		WaitingFor: wait.ForHTTP("/").WithPort("8529/tcp").WithStartupTimeout(2 * time.Minute),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("Failed to start ArangoDB container: %v", err)
	}

	// Ensure the container is terminated after tests
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			log.Fatalf("Failed to terminate ArangoDB container: %v", err)
		}
	}()

	host, err := container.Host(ctx)
	if err != nil {
		log.Fatalf("Failed to get container host: %v", err)
	}
	mappedPort, err := container.MappedPort(ctx, "8529")
	if err != nil {
		log.Fatalf("Failed to get mapped port: %v", err)
	}

	// Prepare ArangoDB configuration
	arangoConf := &config.ArangoConfig{
		User:               "root",
		Pass:               "rootpassword",
		ConnStrs:           fmt.Sprintf("http://%s:%s", host, mappedPort.Port()),
		DBName:             "test",
		InsecureSkipVerify: true,
	}

	t.Run("Test arango db initialization", func(t *testing.T) {
		_, _, err := createSyncDatabaseAndCollections(ctx, "test", "videos_collection", arangoConf.ConnStrs)
		require.NoError(t, err)

		conn, err := NewArangoDB(ctx, arangoConf)
		require.NoError(t, err)

		assert.IsType(t, &arangoDB{}, conn)
	})

	t.Run("Test return database instance", func(t *testing.T) {
		ctx := context.Background()
		conn, err := NewArangoDB(ctx, arangoConf)
		require.NoError(t, err)

		db := conn.Database(ctx)

		assert.NotNil(t, db)
	})

	t.Run("Test get collection", func(t *testing.T) {
		ctx := context.Background()
		conn, err := NewArangoDB(ctx, arangoConf)
		require.NoError(t, err)

		col, err := conn.GetCollection(ctx, "videos_collection")
		require.NoError(t, err)

		assert.Equal(t, col.Name(), "videos_collection")
		assert.NotNil(t, col)
	})
}

// helper functions
func createSyncDatabaseAndCollections(ctx context.Context, dbName string, colName string, connStr string) (arangodb.Database, arangodb.Collection, error) {
	endpoint := connection.NewRoundRobinEndpoints([]string{connStr})
	conn := connection.NewHttp2Connection(connection.DefaultHTTP2ConfigurationWrapper(endpoint /*InsecureSkipVerify*/, true))
	err := conn.SetAuthentication(connection.NewBasicAuth("root", "rootpassword"))
	if err != nil {
		return nil, nil, err
	}
	client := arangodb.NewClient(conn)
	db, err := client.CreateDatabase(ctx, dbName, nil)
	if err != nil {
		return nil, nil, err
	}

	col, err := db.CreateCollection(ctx, colName, nil)
	if err != nil {
		return nil, nil, err
	}

	return db, col, nil
}
