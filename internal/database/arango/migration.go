package arango

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mashaghel/internal/config"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/arangodb/go-driver/v2/arangodb"
)

type ArangoMigration interface {
	CreateFile(path string, fileName string) error
	Apply(path string, version string) error
	Rollback(path string, colName string) error
}

type arangoMigration struct {
	db     arangodb.Database
	config *config.ArangoConfig
}

func NewMigration(db arangodb.Database, config *config.ArangoConfig) ArangoMigration {
	return &arangoMigration{
		db:     db,
		config: config,
	}
}

func (a *arangoMigration) CreateFile(path string, fileName string) error {
	// Directory of migrations
	migrationDir, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Create migrations directory if not exists
	if _, err := os.Stat(migrationDir); os.IsNotExist(err) {
		os.Mkdir(migrationDir, os.ModePerm)
		if !os.IsNotExist(err) {
			return err
		}
	}

	// Json file title creation
	fileFullName := fmt.Sprintf("%d_%s.json", time.Now().Unix(), fileName)

	// Create json file and write template
	file, err := os.Create(filepath.Join(migrationDir, fileFullName))
	if err != nil {
		return err
	}
	fileTemplate, err := generateTemplate()
	if err != nil {
		return err
	}
	_, err = file.Write(fileTemplate)
	if err != nil {
		return err
	}
	defer file.Close()

	log.Printf("Migration file %s created successfully", fileFullName)
	return nil
}

func (a *arangoMigration) Apply(path string, version string) error {
	ctx := context.Background()
	db := a.db

	// Create migrations_record collection if does not exist
	migrationsExist, err := migrationsCollectionExists(ctx, db)
	if err != nil {
		return err
	}
	if !migrationsExist {
		_, err = createMigrationsCollection(ctx, db)
		log.Println("Migrations collection created")
		if err != nil {
			return err
		}
	}

	// Get migration versions
	versions, err := getVersions(ctx, db)
	if err != nil {
		return err
	}

	// Get migration files versions
	migrationFiles, err := getMigrationFiles(path)
	if err != nil {
		return err
	}

	if len(migrationFiles) == 0 {
		return fmt.Errorf("no migration files found in %s", path)
	}

	if version != "" {
		if slices.Contains(migrationFiles, version) && !slices.Contains(versions, version) {
			migrationConf, err := readMigrationFile(version, path)
			if err != nil {
				return err
			}
			collectionConf := migrationConf.Up

			// Check if collection exists
			exists, err := db.CollectionExists(ctx, collectionConf.CollectionName)
			if err != nil {
				return err
			}
			// If collection exists, update it
			if exists {
				//update collection
				err = updateCollection(ctx, db, collectionConf, version)
				if err != nil {
					return err
				}

				log.Printf("Collection %s updated successfully", collectionConf.CollectionName)
			} else {
				// If collection does not exist, create it
				collection, err := createCollection(ctx, db, version, path)
				if err != nil {
					return err
				}

				log.Printf("Collection %s created successfully", collection.Name())
			}
		}
	}

	// Get version of the files that are not migrated yet
	var notMigratedFiles []string
	for _, file := range migrationFiles {
		if !slices.Contains(versions, file) {
			notMigratedFiles = append(notMigratedFiles, file)
		}
	}

	for _, file := range notMigratedFiles {
		migrationConf, err := readMigrationFile(file, path)
		if err != nil {
			return err
		}
		collectionConf := migrationConf.Up

		// Check if collection exists
		exists, err := db.CollectionExists(ctx, collectionConf.CollectionName)
		if err != nil {
			return err
		}
		// If collection exists, update it
		if exists {
			//update collection
			err = updateCollection(ctx, db, collectionConf, file)
			if err != nil {
				return err
			}
		} else {
			// If collection does not exist, create it
			collection, err := createCollection(ctx, db, file, path)
			if err != nil {
				return err
			}

			log.Printf("Collection %s created successfully", collection.Name())
		}
	}

	log.Println("Migrations applied successfully")
	return nil
}

func (a *arangoMigration) Rollback(path string, version string) error {
	ctx := context.Background()
	db := a.db

	migrationVersions, err := getVersions(ctx, db)
	if err != nil {
		return err
	}

	if version == "" {
		version = migrationVersions[len(migrationVersions)-1]
		migrationConf, err := readMigrationFile(version, path)
		if err != nil {
			return err
		}
		collectionConf := migrationConf.Down

		rule, ok := collectionConf.Properties.Schema.Rule.(map[string]interface{})
		if !ok {
			return fmt.Errorf("incorrect type of rule")
		}
		// If collection's schema's rule is and empty json, remove the collection
		if len(rule) == 0 {
			collection, err := db.GetCollection(ctx, collectionConf.CollectionName, nil)
			if err != nil {
				log.Println("Failed to get collection", collectionConf.CollectionName)
				return err
			}
			err = collection.Remove(ctx)
			if err != nil {
				log.Println("Failed to remove collection ", collectionConf.CollectionName)
				return err
			}
			err = deleteMigrationRecord(ctx, db, version)
			if err != nil {
				return err
			}
		} else {
			// If collection's schema's rule is not empty, update the collection
			err = updateCollection(ctx, db, collectionConf, version)
			if err != nil {
				log.Println("Failed to update collection ", collectionConf.CollectionName)
				return err
			}
			err = deleteMigrationRecord(ctx, db, version)
			if err != nil {
				return err
			}
		}
		log.Printf("Collection %s updated", collectionConf.CollectionName)
		log.Println("Migrations rolled back successfully")
	} else {
		if !slices.Contains(migrationVersions, version) {
			log.Println("Failed to find migration: ", version)
			return fmt.Errorf("no applied migration with this version found: %s", version)
		}

		var rollbackArray []string
		for i, v := range migrationVersions {
			if v == version {
				for j := i + 1; j < len(migrationVersions); j++ {
					rollbackArray = append(rollbackArray, migrationVersions[j])
				}
				break
			}
		}

		for _, element := range rollbackArray {
			migrationConf, err := readMigrationFile(element, path)
			if err != nil {
				return err
			}
			collectionConf := migrationConf.Down

			// If collection's schema's rule is and empty json, remove the collection
			if collectionConf.Properties.Schema.Rule == nil {
				collection, err := db.GetCollection(ctx, collectionConf.CollectionName, nil)
				if err != nil {
					log.Println("Failed to get collection", collectionConf.CollectionName)
					return err
				}
				err = collection.Remove(ctx)
				if err != nil {
					log.Println("Failed to remove collection ", collectionConf.CollectionName)
					return err
				}
				err = deleteMigrationRecord(ctx, db, element)
				if err != nil {
					return err
				}
			} else {
				// If collection's schema's rule is not empty, update the collection
				err = updateCollection(ctx, db, collectionConf, version)
				if err != nil {
					log.Println("Failed to update collection ", collectionConf.CollectionName)
					return err
				}
				err = deleteMigrationRecord(ctx, db, element)
				if err != nil {
					return err
				}
			}
		}

		log.Println("Migrations rolled back to: ", version)
	}

	return nil
}

func createMigrationsCollection(ctx context.Context, db arangodb.Database) (arangodb.Collection, error) {
	// Create collection properties
	cacheEnabled := true
	enforceReplicationFactor := false
	schema := map[string]interface{}{
		"properties": map[string]interface{}{
			"version": map[string]interface{}{
				"type": "string",
			},
		},
		"required": []string{
			"version",
		},
		"additionalProperties": false,
	}
	properties := arangodb.CreateCollectionProperties{
		CacheEnabled: &cacheEnabled,
		Schema: &arangodb.CollectionSchemaOptions{
			Rule:    schema,
			Level:   "strict",
			Message: "Schema for migrations_record collection does not fulfill the requirements.",
		},
	}
	// Create collection options
	options := arangodb.CreateCollectionOptions{
		EnforceReplicationFactor: &enforceReplicationFactor,
	}

	// Create collection
	collection, err := db.CreateCollectionWithOptions(ctx, "migrations_record", &properties, &options)
	if err != nil {
		return nil, err
	}
	return collection, nil
}

func migrationsCollectionExists(ctx context.Context, db arangodb.Database) (bool, error) {
	exists, err := db.CollectionExists(ctx, "migrations_record")
	if err != nil {
		return false, err
	}
	return exists, nil
}

func getVersions(ctx context.Context, db arangodb.Database) ([]string, error) {
	query := `FOR v IN migrations_record RETURN v.version`
	cursor, err := db.Query(ctx, query, nil)
	if err != nil {
		log.Println("Failed to query migrations_record collection")
		return nil, err
	}
	defer cursor.Close()

	var versions []string
	for {
		if !cursor.HasMore() {
			break
		}
		var version string
		_, err := cursor.ReadDocument(ctx, &version)
		if err != nil {
			log.Println("Failed to read document while getting migration versions:", err)
			return nil, err
		}
		versions = append(versions, version)
	}

	return versions, nil
}

func getMigrationFiles(path string) ([]string, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %v", err)
	}

	var jsonFiles []string

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			parts := strings.Split(file.Name(), "_")
			var version string
			if len(parts) > 0 {
				version = parts[0]
			}
			jsonFiles = append(jsonFiles, version)
		}
	}

	return jsonFiles, nil
}

type collectionConfig struct {
	CollectionName string                              `json:"collection_name"`
	Options        arangodb.CreateCollectionOptions    `json:"options"`
	Properties     arangodb.CreateCollectionProperties `json:"properties"`
}
type migration struct {
	Up   collectionConfig
	Down collectionConfig
}

func generateTemplate() ([]byte, error) {
	log.Println("Generating migration file ...")
	enforceReplicationFactor := true
	config := collectionConfig{
		CollectionName: "collection_name",
		Options: arangodb.CreateCollectionOptions{
			EnforceReplicationFactor: &enforceReplicationFactor,
		},
		Properties: arangodb.CreateCollectionProperties{
			CacheEnabled:          nil,
			DistributeShardsLike:  "",
			DoCompact:             nil,
			IndexBuckets:          16,
			InternalValidatorType: 0,
			IsDisjoint:            false,
			IsSmart:               false,
			IsSystem:              false,
			IsVolatile:            false,
			JournalSize:           1048576,
			KeyOptions:            nil,
			MinReplicationFactor:  1,
			NumberOfShards:        1,
			ReplicationFactor:     1,
			Schema: &arangodb.CollectionSchemaOptions{
				Rule:    nil,
				Level:   arangodb.CollectionSchemaLevelModerate,
				Message: "Schema of collection_name collection does not fulfill the requirements.",
			},
			ShardingStrategy:    "",
			ShardKeys:           []string{"_key"},
			SmartGraphAttribute: "",
			SmartJoinAttribute:  "",
			SyncByRevision:      false,
			Type:                arangodb.CollectionTypeDocument,
			WaitForSync:         true,
			WriteConcern:        1,
			ComputedValues:      nil,
		},
	}

	migration := migration{
		Up:   config,
		Down: config,
	}

	byteFile, err := json.MarshalIndent(&migration, "", "  ")
	if err != nil {
		log.Println("Failed to convert json to byte:", err)
		return nil, err
	}

	return byteFile, nil
}

func createCollection(ctx context.Context, db arangodb.Database, version string, path string) (arangodb.Collection, error) {
	migrationConf, err := readMigrationFile(version, path)
	if err != nil {
		return nil, err
	}
	collectionConf := migrationConf.Up

	options := arangodb.CreateCollectionOptions{
		EnforceReplicationFactor: collectionConf.Options.EnforceReplicationFactor,
	}

	properties := arangodb.CreateCollectionProperties{
		CacheEnabled:          collectionConf.Properties.CacheEnabled,
		DistributeShardsLike:  collectionConf.Properties.DistributeShardsLike,
		DoCompact:             collectionConf.Properties.DoCompact,
		IndexBuckets:          collectionConf.Properties.IndexBuckets,
		InternalValidatorType: collectionConf.Properties.InternalValidatorType,
		IsDisjoint:            collectionConf.Properties.IsDisjoint,
		IsSmart:               collectionConf.Properties.IsSmart,
		IsSystem:              collectionConf.Properties.IsSystem,
		IsVolatile:            collectionConf.Properties.IsVolatile,
		JournalSize:           collectionConf.Properties.JournalSize,
		KeyOptions:            collectionConf.Properties.KeyOptions,
		NumberOfShards:        collectionConf.Properties.NumberOfShards,
		ReplicationFactor:     collectionConf.Properties.ReplicationFactor,
		Schema:                collectionConf.Properties.Schema,
		ShardingStrategy:      collectionConf.Properties.ShardingStrategy,
		ShardKeys:             collectionConf.Properties.ShardKeys,
		SmartGraphAttribute:   collectionConf.Properties.SmartGraphAttribute,
		SmartJoinAttribute:    collectionConf.Properties.SmartJoinAttribute,
		SyncByRevision:        collectionConf.Properties.SyncByRevision,
		Type:                  collectionConf.Properties.Type,
		WaitForSync:           collectionConf.Properties.WaitForSync,
		WriteConcern:          collectionConf.Properties.WriteConcern,
		ComputedValues:        collectionConf.Properties.ComputedValues,
	}

	collection, err := db.CreateCollectionWithOptions(ctx, collectionConf.CollectionName, &properties, &options)
	if err != nil {
		return nil, err
	}

	collection.Properties(ctx)

	key, err := addMigrationRecord(ctx, db, version)
	if err != nil {
		err = deleteMigrationRecord(ctx, db, key)
		if err != nil {
			return nil, err
		}
		return nil, err
	}

	return collection, nil
}

func readMigrationFile(version string, path string) (migration, error) {
	// Get and read the file content we want to apply
	files, err := os.ReadDir(path)
	if err != nil {
		return migration{}, fmt.Errorf("failed to read directory: %v", err)
	}

	var byteContent []byte
	var migrationConf migration
	var fileName string
	for _, file := range files {
		if strings.Contains(file.Name(), version) {
			fileName = file.Name()
			byteContent, err = os.ReadFile(filepath.Join(path, file.Name()))
			if err != nil {
				log.Printf("Failed to read file %s: %v", file.Name(), err)
				return migration{}, err
			}
			break
		}
	}

	err = json.Unmarshal(byteContent, &migrationConf)
	if err != nil {
		log.Printf("Failed to convert %s to collectionConfig: %v", fileName, err)
		return migration{}, err
	}

	return migrationConf, nil
}

func addMigrationRecord(ctx context.Context, db arangodb.Database, version string) (string, error) {
	collection, err := db.GetCollection(ctx, "migrations_record", nil)
	if err != nil {
		return "", err
	}

	doc, err := collection.CreateDocument(ctx, map[string]interface{}{
		"version": version,
	})
	if err != nil {
		return "", err
	}

	return doc.Key, nil
}

func deleteMigrationRecord(ctx context.Context, db arangodb.Database, version string) error {
	_, err := db.GetCollection(ctx, "migrations_record", nil)
	if err != nil {
		return err
	}

	query := `
		FOR doc IN migrations_record
			FILTER doc.version == @version
			REMOVE doc IN migrations_record
			RETURN OLD
	`
	opts := arangodb.QueryOptions{
		BindVars: map[string]interface{}{
			"version": version,
		},
	}

	cursor, err := db.Query(ctx, query, &opts)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer cursor.Close()

	// Check if the document was deleted
	if !cursor.HasMore() {
		return fmt.Errorf("no document found with version: %s", version)
	}

	return nil
}

func updateCollection(ctx context.Context, db arangodb.Database, collectionConf collectionConfig, version string) error {
	properties := arangodb.SetCollectionPropertiesOptions{
		WaitForSync:       &collectionConf.Properties.WaitForSync,
		JournalSize:       collectionConf.Properties.JournalSize,
		ReplicationFactor: collectionConf.Properties.ReplicationFactor,
		WriteConcern:      collectionConf.Properties.WriteConcern,
		CacheEnabled:      collectionConf.Properties.CacheEnabled,
		Schema:            collectionConf.Properties.Schema,
		ComputedValues:    collectionConf.Properties.ComputedValues,
	}

	collection, err := db.GetCollection(ctx, collectionConf.CollectionName, nil)
	if err != nil {
		log.Printf("Error getting collection: %s", err)
		return err
	}

	err = collection.SetProperties(ctx, properties)
	if err != nil {
		log.Printf("Error updating collection properties: %s", err)
		return err
	}

	_, err = addMigrationRecord(ctx, db, version)
	if err != nil {
		err = deleteMigrationRecord(ctx, db, version)
		if err != nil {
			return err
		}
		return err
	}

	return nil
}
