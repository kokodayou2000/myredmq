package cmd

import (
	"github.com/spf13/cobra"
)

// keyCmd represents the key command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create something",
	Long:  `you can create topic <topic> or  create consumer <consumer> or create group <groupID>`,

	Run: func(cmd *cobra.Command, args []string) {
		//createMag := services.GetCreateMag()

		if cmd.Flag("topic").Value.String() == "true" {
			if len(args) == 0 {
				cmd.Help()
				return
			}
			cmd.Println("create topic ", args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().BoolP("topic", "t", false, "set topic")
}
