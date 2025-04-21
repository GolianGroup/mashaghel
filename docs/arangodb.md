# 📚 ArangoDB

To work with ArangoDB in this application, follow these steps:

## Migrations

1. Generate a new JSON migration file if you are creating a new collection or making changes to a collection:

```bash
go run main.go ag_makemigration
go run main.go ag_makemigration --dir ./internal/database/arango/migrations
```

- Change collection_name in the file to the desired collection name.
- Up Migration Changes:

Add a rule section to the schema or update the existing one with your fields and validations.
If updating an existing schema, provide the complete new schema.

- Down Migration Changes:

For a new collection, write an empty rule object.
If updating an existing schema, jot down the complete previous schema in the rule section.

2. Apply Migration Changes:

```bash
go run main.go ag_migrate
go run main.go ag_migrate --dir ./internal/database/arango/migrations
go run main.go ag_migrate --version 12345
```

This will create a new collection or apply the changes and add a new migration record to the migrations_record collection.

4. Rollback Migrations:

```bash
go run main.go ag_rollback
go run main.go ag_rollback --dir ./internal/database/arango/migrations
go run main.go ag_rollback --version 12345
```

This rolls back migrations up to, but not including, the specified version.

## Repository

1. To access a collection:

Add it to ./internal/database/arango/migration.go.

2. Utilize Arango’s built-in CRUD functions, knowing they use the \_key field. If avoiding the \_key field, write raw AQL queries.
   In the repository layer:

Develop a struct with complete validation for the object you are inserting.
Create another struct for the object you are fetching.
Place the inserting struct in the dto directory and the fetching object in the dao directory.
Align your DTO with the schema of the collection carefully.
