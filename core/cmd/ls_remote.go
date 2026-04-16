package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/whimsy/jdm/internal/jdk"
)

var lsRemoteCmd = &cobra.Command{
	Use:   "ls-remote [version]",
	Short: "List available JDK versions from remote",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vm := jdk.NewVersionManager(Cfg)

		version := "17" // default
		if len(args) > 0 {
			version = args[0]
		}

		versions, err := vm.ListRemote(version)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(versions) == 0 {
			fmt.Printf("No versions found for Java %s\n", version)
			return
		}

		fmt.Printf("Available JDK versions (Java %s):\n", version)
		for _, v := range versions {
			lts := ""
			if isLTS(v.ReleaseName) {
				lts = " [LTS]"
			}
			fmt.Printf("  %s%s\n", v.Version, lts)
		}
	},
}

func isLTS(releaseName string) bool {
	return len(releaseName) >= 3 && releaseName[len(releaseName)-3:] == "LTS"
}

func init() {
	rootCmd.AddCommand(lsRemoteCmd)
}