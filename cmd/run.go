package cmd

import (
	"fmt"
	"myredmq/myredmq/redis"
	"myredmq/myredmq/util"
	"myredmq/services"

	"github.com/spf13/cobra"
)

// Cfg 创建对象
var Cfg util.Config

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run service",
	Short: "Run",
	Run: func(cmd *cobra.Command, args []string) {
		// load config file execute file must some level
		err := Cfg.LoadConfig(".")
		if err != nil {
			return
		}
		Run()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func Run() {
	fmt.Print("Welcome to the redisMQ.  \nType 'exit' to quit. \n")
	client := redis.NewClient(Cfg.Redis.Network, Cfg.Redis.Address, Cfg.Redis.Password)
	if client != nil {
		fmt.Println("connect redis success!")
	}
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
