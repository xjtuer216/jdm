package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/whimsy/jdm/internal/jdk"
)

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the currently active JDK version",
	Run: func(cmd *cobra.Command, args []string) {
		vm := jdk.NewVersionManager(Cfg)
		current, err := vm.GetCurrent()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if current == nil {
			fmt.Println("No JDK version currently active.")
			fmt.Println("Run 'jdm use <version>' to activate a version.")
			return
		}

		fmt.Printf("Current JDK: %s\n", current.Version)
		fmt.Printf("Path: %s\n", current.Path)
	},
}

func init() {
	rootCmd.AddCommand(currentCmd)
}