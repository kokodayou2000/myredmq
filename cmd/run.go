package cmd

import (
	"fmt"
	"myredmq/myredmq/util"
	"myredmq/services"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run service",
	Short: "Run",
	Run: func(cmd *cobra.Command, args []string) {
		// load config file execute file must some level
		config, err := util.LoadConfig(".")
		if err != nil {
			return
		}
		fmt.Println(config.Redis.Address)

		Run()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func Run() {
	fmt.Print("Welcome to the redisMQ. \nType 'exit' to quit. \n")
	for {
		question := services.AskUserQuestion()
		if question == "exit" {
			break
		}
		res, ok := services.AnswerQuestion(question)
		if ok {
			fmt.Print(res)
			fmt.Println()
		}
	}
}
