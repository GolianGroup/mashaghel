/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"mashaghel/internal/config"
	"mashaghel/internal/database/arango"

	"github.com/spf13/cobra"
)

// arangoMigrateCmd represents the ag_migrate command
var arangoMigrateCmd = &cobra.Command{
	Use:   "ag_migrate",
	Short: "Migrate the json files",
	Long: `Migrate the migration files. Example:
	ag_migrate                              Migrate all the migration files
	ag_migrate --dir ./database/arangomigrations  Migrate all the migration files from sepecific directory
	ag_migrate --version  123456                 Migrate the migration file with this hash`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("ag_makemigration called")

		dirFlag, err := cmd.Flags().GetString("dir")
		if err != nil {
			cmd.PrintErrf("Error while getting dir flag: %s", err.Error())
			return
		}

		versionFlag, err := cmd.Flags().GetString("version")
		if err != nil {
			cmd.PrintErrf("Error while getting version flag: %s", err.Error())
			return
		}

		dbConfig, err := config.LoadConfig("config/config.yml")
		if err != nil {
			log.Panicf("failed to setup viper: %s", err)
			return
		}

		ctx := cmd.Context()

		db, err := arango.NewArangoDB(ctx, &dbConfig.ArangoDB)
		if err != nil {
			cmd.PrintErrf("failed to setup arango for migrations: %s", err)
			return
		}

		migration := arango.NewMigration(db.Database(ctx), &dbConfig.ArangoDB)
		err = migration.Apply(dirFlag, versionFlag)
		if err != nil {
			cmd.PrintErrf("Error while applying migration:\n\t %v", err)
			return
		}
	},
}

func init() {
	RootCmd.AddCommand(arangoMigrateCmd)
	arangoMigrateCmd.Flags().String("dir", "./internal/database/arango/migrations", "Directory of the migrations")
	arangoMigrateCmd.Flags().String("version", "", "Version of the migration that is going to be applied")
}
