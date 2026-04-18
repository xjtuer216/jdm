package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xjtuer216/jdm/internal/jdk"
)

var installCmd = &cobra.Command{
	Use:   "install <version>",
	Short: "Install a JDK version",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vm := jdk.NewVersionManager(getConfig())
		version := args[0]
		fmt.Printf("Installing JDK %s...\n", version)
		if err := vm.Install(version); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
