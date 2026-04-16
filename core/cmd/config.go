package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long: `Manage JDM configuration.
Configuration can be modified via command or by editing config.json directly.

Config file locations:
  - User config: ~/.jdm/config.json
  - Install config: <exe-dir>/config.json

The install config serves as a template for first-time users.`,
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := getConfig()
		key := args[0]
		value := cfg.Get(key)
		if value == "" {
			fmt.Printf("Key '%s' not found\n", key)
			os.Exit(1)
		}
		fmt.Printf("%s: %s\n", key, value)
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := getConfig()
		key := args[0]
		value := args[1]

		if err := cfg.Set(key, value); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("%s set to '%s'\n", key, value)
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := getConfig()
		all := cfg.GetAll()

		fmt.Println("Current configuration:")
		for key, value := range all {
			if key == "default" && value == "" {
				value = "(not set)"
			}
			fmt.Printf("  %s: %s\n", key, value)
		}

		// 显示配置文件的路径
		fmt.Printf("\nUser config: %s\n", cfg.GetConfigPath())
		if installPath := cfg.GetInstallConfigPath(); installPath != "" {
			fmt.Printf("Install config: %s\n", installPath)
		}
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration file",
	Long: `Create a default configuration file.
This will create config.json in the user directory (~/.jdm/config.json).`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := getConfig()
		if err := cfg.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Configuration saved to: %s\n", cfg.GetConfigPath())
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configInitCmd)

	// 注册配置项补全
	_ = configGetCmd.RegisterFlagCompletionFunc("key", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"jdm_home", "jdk_home", "mirror", "default"}, cobra.ShellCompDirectiveNoFileComp
	})
	_ = configSetCmd.RegisterFlagCompletionFunc("key", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"jdm_home", "jdk_home", "mirror", "default"}, cobra.ShellCompDirectiveNoFileComp
	})
}