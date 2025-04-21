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

// arangoRollbackCmd represents the ag_rollback command
var arangoRollbackCmd = &cobra.Command{
	Use:   "ag_rollback",
	Short: "Rollback migration/migrations",
	Long: `Rollback migration migration/migrations. The path should be path to migration files.
		You can write the migration version to rollback to a specific version.
		For example:
		ag_rollback
		ag_rollback --dir ./database/arango/migrations --version 12345`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("ag_rollback called")

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
		err = migration.Rollback(dirFlag, versionFlag)
		if err != nil {
			cmd.PrintErrf("Error while rolling migration back:\n\t %v", err)
			return
		}
	},
}

func init() {
	RootCmd.AddCommand(arangoRollbackCmd)
	arangoRollbackCmd.Flags().String("dir", "./internal/database/arango/migrations", "Directory of the migrations")
	arangoRollbackCmd.Flags().String("version", "", "Version of the migration that migrations will be rolled back to")
}
