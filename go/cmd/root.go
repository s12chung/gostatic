/*
Package cmd is the CLI cmd interface for the gostatic binary (gostatic init, etc.)
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gostatic",
	Short: "Static site generator for web developers",
	Long: `Use Go and Webpack to generate static websites like a standard Go web application.
You can even run gostatic apps, as a web app.
See https://github.com/s12chung/gostatic for more.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Command without args does not exist yet!")
	},
}

var test bool

func init() {
	rootCmd.PersistentFlags().BoolVarP(&test, "test", "t", false, "Running in test mode")
}

// Execute starts the CLI cmd prompt program
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
