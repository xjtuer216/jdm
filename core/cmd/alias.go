package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var aliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Manage version aliases",
}

var aliasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all aliases",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := getConfig()
		if len(cfg.Aliases) == 0 {
			fmt.Println("No aliases defined.")
			return
		}

		fmt.Println("Aliases:")
		for name, version := range cfg.Aliases {
			fmt.Printf("  %s -> %s\n", name, version)
		}
	},
}

var aliasSetCmd = &cobra.Command{
	Use:   "set <name> <version>",
	Short: "Set an alias for a version",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := getConfig()
		name := args[0]
		version := args[1]

		cfg.SetAlias(name, version)
		if err := cfg.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Alias '%s' set to '%s'\n", name, version)
	},
}

var aliasDelCmd = &cobra.Command{
	Use:   "del <name>",
	Short: "Delete an alias",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := getConfig()
		name := args[0]

		if _, ok := cfg.Aliases[name]; !ok {
			fmt.Fprintf(os.Stderr, "Alias '%s' does not exist\n", name)
			os.Exit(1)
		}

		cfg.RemoveAlias(name)
		if err := cfg.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Alias '%s' deleted\n", name)
	},
}

func init() {
	rootCmd.AddCommand(aliasCmd)
	aliasCmd.AddCommand(aliasListCmd)
	aliasCmd.AddCommand(aliasSetCmd)
	aliasCmd.AddCommand(aliasDelCmd)
}