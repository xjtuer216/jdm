package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xjtuer216/jdm/internal/config"
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
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return config.ReadableKeys(), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		cfg := getConfig()
		warnDeprecated(args[0])
		key := config.CanonicalKey(args[0])

		value, err := cfg.Get(key)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
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
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return config.SettableKeys(), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		cfg := getConfig()
		warnDeprecated(args[0])
		key := config.CanonicalKey(args[0])
		value := args[1]

		if err := cfg.Set(key, value); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("%s set to '%s'\n", key, value)
	},
}

// warnDeprecated 在使用已弃用的配置项名称时输出提示
func warnDeprecated(key string) {
	if !config.IsDeprecatedKey(key) {
		return
	}
	fmt.Fprintf(os.Stderr, "Warning: '%s' is deprecated, use '%s' instead\n", key, config.CanonicalKey(key))
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := getConfig()
		all, err := cfg.GetAll()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Current configuration:")
		for key, value := range all {
			if key == "default" && value == "" {
				value = "(not set)"
			}
			fmt.Printf("  %s: %s\n", key, value)
		}

		// 显示配置文件的路径
		fmt.Printf("\nUser config: %s\n", cfg.GetConfigPath())
		if installPath := config.InstallConfigPath(); installPath != "" {
			fmt.Printf("Install config: %s\n", installPath)
		}
	},
}

var configInitInstall bool

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration file",
	Long: `Create a default configuration file.
This will create config.json in the user directory (~/.jdm/config.json).

With --install, initialize the install-directory config template instead:
missing keys are filled with defaults, existing values are kept.
The Windows installer uses this flag.`,
	Run: func(cmd *cobra.Command, args []string) {
		if configInitInstall {
			// 不加载用户配置：安装目录模板的默认值以 exe 所在目录为准
			installPath := config.InstallConfigPath()
			if installPath == "" {
				fmt.Fprintln(os.Stderr, "Error: cannot locate the install directory")
				os.Exit(1)
			}
			if err := config.EnsureInstallConfig(installPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Install config initialized at: %s\n", installPath)
			return
		}

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
	configInitCmd.Flags().BoolVar(&configInitInstall, "install", false, "initialize the install-directory config template")
}
