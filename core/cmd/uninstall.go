package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/whimsy/jdm/internal/jdk"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <version>",
	Short: "Uninstall a JDK version",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vm := jdk.NewVersionManager(getConfig())
		version := args[0]
		fmt.Printf("Uninstalling JDK %s...\n", version)
		if err := vm.Uninstall(version); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}