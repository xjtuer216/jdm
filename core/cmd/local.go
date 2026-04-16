package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var localCmd = &cobra.Command{
	Use:   "local <version>",
	Short: "Set project-local JDK version (reserved for future)",
	Long: `This command is reserved for future implementation.
It will create a .jdmrc file in the project directory for automatic version switching.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("This feature is not yet implemented.")
		fmt.Println("It will be available in a future release.")
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(localCmd)
}