package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var GitCommitID string

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version display the 40-byte hexadecimal name of the last git commit object",

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Git Commit:", GitCommitID)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
