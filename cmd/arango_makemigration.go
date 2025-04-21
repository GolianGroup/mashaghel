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

// arangoMakemigrationCmd represents the arangoMakemigration command
var arangoMakemigrationCmd = &cobra.Command{
	Use:   "ag_makemigration [name]",
	Short: "Create a new json migration file for writing schemas",
	Long: `ag_makemigration add_users_collection // Create a new json migration file
		ag_makemigration add_users_collection --dir ./database/arango/migrations // Create a new json migration file in specific directory`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("ag_makemigration called")

		name := args[0]
		if name == "" {
			cmd.PrintErr("Please provide a name for the migration")
			return
		}

		dirFlag, err := cmd.Flags().GetString("dir")
		if err != nil {
			cmd.PrintErrf("Error while getting dir flag: %s", err.Error())
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
		err = migration.CreateFile(dirFlag, name)
		if err != nil {
			cmd.PrintErrf("Error while creating migration file:\n\t %v", err)
			return
		}
	},
}

func init() {
	RootCmd.AddCommand(arangoMakemigrationCmd)

	arangoMakemigrationCmd.Flags().String("dir", "./internal/database/arango/migrations", "Directory of arango migrations")
}
