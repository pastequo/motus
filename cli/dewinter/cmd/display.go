package cmd

import (
	"github.com/pastequo/motus"
	"github.com/spf13/cobra"
)

var txt string
var okCount int
var amissCount int

// displayCmd represents the display command
var displayCmd = &cobra.Command{
	Use:   "display",
	Short: "Display the input txt, lingo style",

	RunE: func(cmd *cobra.Command, args []string) error {

		return motus.DisplayText(txt, okCount, amissCount)
	},
}

func init() {

	displayCmd.Flags().StringVarP(&txt, "txt", "t", "", "input text")
	displayCmd.MarkFlagRequired("txt")

	displayCmd.Flags().IntVarP(&okCount, "okCount", "o", 0, "number of character correctly placed")
	displayCmd.MarkFlagRequired("okCount")

	displayCmd.Flags().IntVarP(&amissCount, "amissCount", "a", 0, "number of character correct nut not at their right place")
	displayCmd.MarkFlagRequired("amissCount")

	rootCmd.AddCommand(displayCmd)
}
