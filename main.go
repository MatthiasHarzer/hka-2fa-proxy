package main

import (
	"fmt"

	"github.com/MatthiasHarzer/hka-2fa-proxy/commands/run"
	"github.com/spf13/cobra"
)

var version = "unknown"

func init() {
	command.AddCommand(run.Command)
}

var command = &cobra.Command{
	Use:   "hka-2fa-proxy",
	Short: "A proxy into the outlook instance of the HKA",
	Long:  "A proxy into the outlook instance of the HKA",
	Run: func(c *cobra.Command, args []string) {
		fmt.Println("hka-2fa-proxy", version)
	},
}

func main() {
	err := command.Execute()
	if err != nil {
		panic(err)
	}
}
