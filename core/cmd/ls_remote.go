package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xjtuer216/jdm/internal/jdk"
)

var lsRemoteCmd = &cobra.Command{
	Use:   "ls-remote [version]",
	Short: "List available JDK versions from remote",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vm := jdk.NewVersionManager(Cfg)

		var version string
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

		if len(args) == 0 {
			// Table format for all major versions
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintf(w, "VERSION\tMAJOR\tLTS\tVENDOR\n")
			for _, v := range versions {
				major := ""
				parts := strings.Split(v.Version, ".")
				if len(parts) > 0 {
					major = parts[0]
				}
				lts := ""
				if v.IsLTS {
					lts = "Yes"
				}
				vendor := "Eclipse Adoptium"
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", v.Version, major, lts, vendor)
			}
			w.Flush()
		} else {
			// Flat list for specific version
			fmt.Printf("Available JDK versions (Java %s):\n", version)
			for _, v := range versions {
				lts := ""
				if v.IsLTS {
					lts = " [LTS]"
				}
				fmt.Printf("  %s%s\n", v.Version, lts)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(lsRemoteCmd)
}
