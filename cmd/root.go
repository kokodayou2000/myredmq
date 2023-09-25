package cmd

import (
	"github.com/spf13/cobra"
	"myredmq/services"
	"os"
)

const (
	Version = "0.0.1"
	AppName = "redisMQ"
)

var createMag *services.CreateMag

var rootCmd = &cobra.Command{
	Use:   AppName,
	Short: "Redis MQ Example",
	Run: func(cmd *cobra.Command, args []string) {
		if flag := cmd.Flag("version"); flag != nil && flag.Value.String() == "true" {
			cmd.Println(AppName, "v"+Version)
		} else {
			cmd.Help()
		}
	},
}

func Execute() {
	err := rootCmd.Execute()

	if err != nil {
		os.Exit(1)
	}
}

func init() {
	services.NewCreateMag()
	rootCmd.Flags().BoolP("version", "v", false, "version of "+AppName)
}
