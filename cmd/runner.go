/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/ocontest/backend/pkg"
	"github.com/ocontest/backend/pkg/configs"
	"github.com/ocontest/backend/internal/runner/consumer"
	"github.com/spf13/cobra"
	"log"
)

// runnerCmd represents the runner command
var runnerCmd = &cobra.Command{
	Use:   "runner",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		configs.InitConf()
		c := configs.Conf
		pkg.InitLog(c.Log)
		pkg.Log.Info("config and log modules initialized")

		RunRunnerTaskHandler(c)
	},
}

func init() {
	rootCmd.AddCommand(runnerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runnerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runnerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func RunRunnerTaskHandler(c *configs.OContestConf) {
	runnerHandler, err := consumer.NewRunnerScheduler(c.Judge.Nats)
	if err != nil {
		log.Fatal("error on creating runner scheduler: ", err)
	}
	runnerHandler.StartListen()
}
