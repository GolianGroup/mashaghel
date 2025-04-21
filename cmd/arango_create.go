/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"log"
	"mashaghel/internal/config"
	"time"

	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/connection"
	"github.com/spf13/cobra"
)

// arangoCreateCmd represents the arangoCreate command
var arangoCreateCmd = &cobra.Command{
	Use:   "ag_create",
	Short: "Creates arangodb database.",
	Long: `ag_create creates database based on environment variables
	ag_create --db_name=mydb creates database with name mydb`,
	Run: func(cmd *cobra.Command, args []string) {
		dbConfig, err := config.LoadConfig("config/config.yml")
		if err != nil {
			log.Panicf("failed to setup viper: %s", err)
			return
		}

		connStrs, err := config.GetArangoStrings(&dbConfig.ArangoDB)
		if err != nil {
			cmd.PrintErrf("failed to setup arango to create database: %s", err)
		}
		endpoint := connection.NewRoundRobinEndpoints(connStrs)
		conn := connection.NewHttp2Connection(connection.DefaultHTTP2ConfigurationWrapper(endpoint /*InsecureSkipVerify*/, dbConfig.ArangoDB.InsecureSkipVerify))

		auth := connection.NewBasicAuth(dbConfig.ArangoDB.User, dbConfig.ArangoDB.Pass)
		err = conn.SetAuthentication(auth)
		if err != nil {
			cmd.PrintErrf("failed to authenticate arango to create database: %s", err)
			return
		}

		client := arangodb.NewClient(conn)

		dbFlag, err := cmd.Flags().GetString("db_name")
		if err != nil {
			cmd.PrintErrf("failed to get database name: %s", err)
			return
		}

		var dbName string
		if dbFlag == "" {
			dbName = dbConfig.ArangoDB.DBName
		} else {
			dbName = dbFlag
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		dbExists, err := client.DatabaseExists(ctx, dbName)
		if err != nil {
			cmd.PrintErrf("failed to check if database exists: %s", err)
			return
		}
		if !dbExists {
			_, err = client.CreateDatabase(ctx, dbName, nil)
			if err != nil {
				cmd.PrintErrf("failed to create database: %s", err)
				return
			}
			cmd.Printf("Database %s created successfully.\n", dbName)
			return
		} else {
			cmd.Printf("Database %s already exists.\n", dbName)
			return
		}
	},
}

func init() {
	RootCmd.AddCommand(arangoCreateCmd)
	arangoCreateCmd.Flags().String("db_name", "", "Database name")
}
