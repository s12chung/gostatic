package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gostatic",
	Short: "Static site generator for web developers",
	Long: `Use Go and Webpack to generate static websites using
web developer patterns like mapping routes to functions.
See https://github.com/s12chung/gostatic for more.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Command without args does not exist yet!")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
