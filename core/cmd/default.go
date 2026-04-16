package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/whimsy/jdm/internal/jdk"
)

var defaultCmd = &cobra.Command{
	Use:   "default <version>",
	Short: "Set the default JDK version",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vm := jdk.NewVersionManager(getConfig())
		version := args[0]
		if err := vm.SetDefault(version); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(defaultCmd)
}