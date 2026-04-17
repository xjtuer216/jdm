package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xjtuer216/jdm/internal/jdk"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List locally installed JDK versions",
	Run: func(cmd *cobra.Command, args []string) {
		vm := jdk.NewVersionManager(Cfg)
		versions, err := vm.ListLocal()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(versions) == 0 {
			fmt.Println("No JDK versions installed.")
			fmt.Println("Run 'jdm install <version>' to install a JDK.")
			return
		}

		fmt.Println("Local JDK versions:")
		for _, v := range versions {
			marker := " "
			if v.IsCurrent {
				marker = "*"
			} else if v.IsDefault {
				marker = ">"
			}
			fmt.Printf("  %s %s (%s)\n", marker, v.Version, v.Path)
		}
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)
}
