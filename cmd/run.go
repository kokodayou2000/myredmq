package cmd

import (
	"context"
	"fmt"
	"myredmq/myredmq/redis"
	"myredmq/myredmq/util"
	"myredmq/services"
	"strings"

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
	if client == nil {
		fmt.Println("connect redis fail!")
		return
	}
	fmt.Println("connect redis success!")
	for {
		question := services.AskUserQuestion()
		if question == "exit" {
			break
		}
		split := strings.Split(question, " ")

		switch strings.ToLower(split[0]) {
		case "create":
			fmt.Println("create ")
			switch strings.ToLower(split[1]) {
			case "topic":

				fmt.Println("topic " + split[2])
			case "group":
				fmt.Println("group " + split[2])
			}
		case "send":
			if len(split) == 4 {
				fmt.Println("send msg " + split[1] + " to " + split[3])
			} else {
				fmt.Println("args error")
			}
		case "set":
			client.Set(context.Background(), split[1], split[2])
		case "get":
			client.Get(context.Background(), split[1])
		}
		res, ok := services.AnswerQuestion(question)
		if ok {
			fmt.Print(res)
			fmt.Println()
		}
	}
}
