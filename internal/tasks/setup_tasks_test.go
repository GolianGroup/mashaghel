package tasks

import (
	"context"
	"fmt"
	"log"
	"mashaghel/internal/config"
	internalCconfig "mashaghel/internal/config"
	"mashaghel/internal/database/scylla"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/panjf2000/ants/v2"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

var scyllaDB scylla.ScyllaDB

func setupScyllaContainer() (scylla.ScyllaDB, testcontainers.Container, error) {
	log.Println("Starting ScyllaDB container")

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "scylladb/scylla:latest",
		ExposedPorts: []string{"9042/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("9042/tcp"),
			wait.ForLog("Starting listening for CQL clients").
				WithStartupTimeout(2*time.Minute),
		),
		Cmd: []string{"--smp", "1"}, // Reduce CPU cores for testing
	}
	scyllaC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start ScyllaDB container: %w", err)
	}

	host, err := scyllaC.Host(ctx)
	if err != nil {
		return nil, scyllaC, fmt.Errorf("failed to get ScyllaDB container host: %w", err)
	}

	port, err := scyllaC.MappedPort(ctx, "9042")
	if err != nil {
		return nil, scyllaC, fmt.Errorf("failed to get ScyllaDB container port: %w", err)
	}
	log.Println("container ready !!!")

	config := &internalCconfig.Config{
		ScyllaDB: internalCconfig.ScyllaDBConfig{
			Hosts:             []string{host + ":" + port.Port()},
			Keyspace:          "test_keyspace",
			Username:          "cassandra",
			Password:          "cassandra",
			ReplicationClass:  "SimpleStrategy",
			ReplicationFactor: 1,
		},
	}

	logger, _ := zap.NewProduction()
	scyllaDb, err := scylla.NewScyllaDB(ctx, config, logger)
	if err != nil {
		return nil, scyllaC, fmt.Errorf("failed to create ScyllaDB session: %w", err)
	}

	return scyllaDb, scyllaC, nil
}

func SetupTables() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Add retry logic for table creation
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		err := createTables(ctx)
		if err == nil {
			break
		}
		if i == maxRetries-1 {
			return fmt.Errorf("failed to create tables after %d retries: %w", maxRetries, err)
		}
		log.Printf("Retry %d: Failed to create tables, waiting before retry: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	return nil
}

func createTables(ctx context.Context) error {
	// scyllaDB tables

	// ########################################################
	// ########################################################
	// ###################      Watch      ####################
	// ########################################################
	// ########################################################

	// Create watched table
	err := scyllaDB.Session().Query(`CREATE TABLE IF NOT EXISTS watched (
		profile_id UUID,
		play_id UUID,
		duration INT,  -- Duration in seconds 
		watched_at TIMEUUID, 
		PRIMARY KEY (profile_id, play_id)
	);`).WithContext(ctx).Exec()
	if err != nil {
		return fmt.Errorf("failed to create watched table: %w", err)
	}

	// Create ordered_watch table
	err = scyllaDB.Session().Query(`CREATE TABLE IF NOT EXISTS ordered_watch (
		profile_id UUID,
		play_id UUID,
		duration INT,
		watched_at TIMEUUID,
		PRIMARY KEY (profile_id, watched_at)
	) WITH CLUSTERING ORDER BY (watched_at DESC);`).WithContext(ctx).Exec()
	if err != nil {
		return fmt.Errorf("failed to create ordered_watch table: %w", err)
	}

	// Create recent_watch table
	err = scyllaDB.Session().Query(`CREATE TABLE IF NOT EXISTS recent_watch (
		profile_id UUID,
		play_id UUID,
		duration INT,  -- Duration in seconds
		watched_at TIMEUUID, 
		PRIMARY KEY (profile_id, play_id)
	)`).WithContext(ctx).Exec()
	if err != nil {
		return fmt.Errorf("failed to create recent_watch table: %w", err)
	}

	// arangoDB collections

	return nil
}

// Function to generate Ids
func generateIDs(count int) []gocql.UUID {
	ids := make([]gocql.UUID, count)
	for i := 0; i < count; i++ {
		ids[i] = gocql.MustRandomUUID()
	}
	return ids
}

func generateData() error {
	query_recent_watch := `
		INSERT INTO recent_watch 
		(profile_id, play_id, duration, watched_at) 
		VALUES (?, ?, ?, ?)
	`

	query_watched := `
		UPDATE watched
		SET duration = ?
		WHERE profile_id = ? AND play_id = ?
	`

	profileIDs := generateIDs(20)
	playIDs := generateIDs(10)

	batch := scyllaDB.Session().NewBatch(gocql.LoggedBatch)
	for i := 0; i < 100; i++ {
		profileID := profileIDs[rand.Intn(len(profileIDs))]
		playID := playIDs[rand.Intn(len(playIDs))]
		duration := (i + 1) * 10 // Example duration
		watchedAt := gocql.UUIDFromTime(time.Now().AddDate(0, 0, -4))

		// Add to batch for recent_watch
		batch.Query(query_recent_watch, profileID, playID, duration, watchedAt)

		// Add to batch for watched
		batch.Query(query_watched, duration, profileID, playID)
	}

	// Execute the batch
	err := scyllaDB.Session().ExecuteBatch(batch)
	if err != nil {
		return fmt.Errorf("failed to execute batch: %v", err)
	}

	return nil
}

func createTaskInstance() *task {
	taskInstance := &task{
		scylla: scyllaDB,
		logger: zap.NewExample(),
		workerpool: func() *ants.Pool {
			pool, err := ants.NewPool(3)
			if err != nil {
				log.Fatalf("failed to create worker pool: %v", err)
			}
			return pool
		}(),
		configs: &config.WorkerPoolConfig{
			TasksConfig: config.TasksConfig{
				WatchCooldownDuration: 3,
				WatchAgeLimit:         24,
			},
		},
	}

	return taskInstance
}

func TestMain(m *testing.M) {
	var (
		scyllaC testcontainers.Container
		err     error
	)

	// Set a longer timeout for container operations
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	scyllaDB, scyllaC, err = setupScyllaContainer()
	if err != nil {
		log.Fatalf("failed to setup ScyllaDB container: %v", err)
	}
	if scyllaC != nil {
		defer func() {
			if err := scyllaC.Terminate(ctx); err != nil {
				log.Printf("failed to terminate ScyllaDB container: %v", err)
			}
		}()
	}

	// Add delay after ScyllaDB container starts
	time.Sleep(5 * time.Second)

	err = SetupTables()
	if err != nil {
		log.Fatalf("failed to setup tables and collections: %v", err)
	}
	err = generateData()
	if err != nil {
		log.Fatalf("failed to insert data to tables: %v", err)
	}

	log.Println("setup finished and tables and collections created, starting tests...")

	// Mock or load configuration explicitly for testing
	// Replace with the actual path to your mock config.yaml

	code := m.Run()

	scyllaDB.Close()

	os.Exit(code)
}
