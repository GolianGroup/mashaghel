package scylla

import (
	"context"
	"fmt"
	"mashaghel/internal/config"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"
)

// ScyllaDB represents the ScyllaDB connection interface
type ScyllaDB interface {
	Session() *gocql.Session
	Ping(ctx context.Context) error
	Close()
}

type scyllaDB struct {
	session *gocql.Session
}

// NewScyllaDB initializes a new ScyllaDB connection
func NewScyllaDB(ctx context.Context, config *config.Config, logger *zap.Logger) (ScyllaDB, error) {
	if len(config.ScyllaDB.Hosts) == 0 {
		return nil, fmt.Errorf("no ScyllaDB hosts provided in configuration")
	}

	// Temporary cluster config to check and create keyspace
	tempCluster := gocql.NewCluster(config.ScyllaDB.Hosts...)
	tempCluster.Authenticator = gocql.PasswordAuthenticator{
		Username: config.ScyllaDB.Username,
		Password: config.ScyllaDB.Password,
	}
	tempCluster.Timeout = 10 * time.Second
	tempCluster.Consistency = gocql.Quorum
	tempCluster.ProtoVersion = 4

	tempSession, err := tempCluster.CreateSession()
	if err != nil {
		logger.Error("Failed to create temporary ScyllaDB session", zap.Error(err))
		return nil, fmt.Errorf("failed to create temporary ScyllaDB session: %w", err)
	}
	defer tempSession.Close()

	// Check if keyspace exists
	var keyspaceName string
	query := "SELECT keyspace_name FROM system_schema.keyspaces WHERE keyspace_name = ?"
	err = tempSession.Query(query, config.ScyllaDB.Keyspace).Scan(&keyspaceName)
	if err != nil && err != gocql.ErrNotFound {
		logger.Error("Failed to check keyspace existence", zap.Error(err))
		return nil, fmt.Errorf("failed to check keyspace existence: %w", err)
	}
	keyspaceExists := keyspaceName != ""

	// If keyspace does not exist, create it
	if !keyspaceExists {
		logger.Info("Keyspace does not exist. Creating keyspace...", zap.String("keyspace", config.ScyllaDB.Keyspace))

		// Use configuration for replication strategy
		replicationClass := config.ScyllaDB.ReplicationClass
		if replicationClass == "" {
			replicationClass = "SimpleStrategy"
		}

		replicationFactor := config.ScyllaDB.ReplicationFactor
		if replicationFactor == 0 {
			replicationFactor = 3
		}

		createKeyspaceQuery := fmt.Sprintf(
			`CREATE KEYSPACE %s WITH replication = {'class': '%s', 'replication_factor': %d}`,
			config.ScyllaDB.Keyspace,
			replicationClass,
			replicationFactor,
		)
		if err := tempSession.Query(createKeyspaceQuery).Exec(); err != nil {
			logger.Error("Failed to create keyspace", zap.Error(err))
			return nil, fmt.Errorf("failed to create keyspace: %w", err)
		}
		logger.Info("Successfully created keyspace", zap.String("keyspace", config.ScyllaDB.Keyspace))
	}

	// Now create the real session with the keyspace
	cluster := gocql.NewCluster(config.ScyllaDB.Hosts...)
	cluster.Keyspace = config.ScyllaDB.Keyspace
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: config.ScyllaDB.Username,
		Password: config.ScyllaDB.Password,
	}
	cluster.Timeout = 10 * time.Second

	// For a single-node development environment, use a lower consistency level
	if config.Environment == "development" {
		cluster.Consistency = gocql.LocalOne
	} else {
		cluster.Consistency = gocql.Quorum // For production
	}

	cluster.RetryPolicy = &gocql.SimpleRetryPolicy{NumRetries: 3}
	cluster.ProtoVersion = 4

	// Configure connection pooling based on expected load
	cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())

	session, err := cluster.CreateSession()
	if err != nil {
		logger.Error("Failed to create ScyllaDB session", zap.Error(err))
		return nil, fmt.Errorf("failed to create ScyllaDB session: %w", err)
	}

	logger.Info("Successfully connected to ScyllaDB", zap.String("keyspace", config.ScyllaDB.Keyspace))
	return &scyllaDB{session: session}, nil
}

// Session returns the ScyllaDB session
func (s *scyllaDB) Session() *gocql.Session {
	return s.session
}

// Ping verifies the ScyllaDB connection
func (s *scyllaDB) Ping(ctx context.Context) error {
	return s.session.Query("SELECT now() FROM system.local").WithContext(ctx).Exec()
}

// Close closes the ScyllaDB session
func (s *scyllaDB) Close() {
	if s.session != nil {
		s.session.Close()
	}
}
