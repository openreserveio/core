package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
}

var rootCmd = &cobra.Command{
	Use:   "coregl",
	Short: "coregl root command",
	Long:  `coregl Collection of Services`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Specify a Subcommand")
		os.Exit(1)
	},
}

func Execute(ctx context.Context) {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
