package cmd

import (
	"context"
	"log"
	app "mashaghel/app"
	"mashaghel/internal/config"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "run",
	Short: "Run Jobs",
	Run:   run,
	Args:  cobra.MaximumNArgs(2),
}

func init() {
	RootCmd.AddCommand(serveCmd)
}

func run(cmd *cobra.Command, args []string) {
	log.Println("run")
	//viper
	config, err := config.LoadConfig("config/config.yml")
	if err != nil {
		log.Fatalf("failed to setup viper: %s", err.Error())
	}
	application := app.NewApplication(context.TODO(), config)
	application.Setup()
}
