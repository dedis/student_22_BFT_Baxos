package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "version of instance",
	Long:  "the version of instance",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("0.0.1 Alpha")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
