package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xjtuer216/jdm/internal/jdk"
)

var useCmd = &cobra.Command{
	Use:   "use <version>",
	Short: "Switch to a different JDK version (current session)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vm := jdk.NewVersionManager(getConfig())
		version := args[0]
		if err := vm.Use(version); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(useCmd)
}
