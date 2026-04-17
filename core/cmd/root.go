package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/xjtuer216/jdm/internal/config"
	"github.com/xjtuer216/jdm/internal/jdk"
	"github.com/xjtuer216/jdm/internal/log"
)

var rootCmd = &cobra.Command{
	Use:   "jdm",
	Short: "JDK Version Manager for Windows",
	Long: `JDM (JDK Version Manager) is a command-line tool for managing multiple JDK versions on Windows.
It helps you install, manage, and switch between different JDK versions quickly and easily.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Cfg is the global config instance
var Cfg *config.Config

// logFile is the path to the log file
var logFile string

func getConfig() *config.Config {
	if Cfg == nil {
		// Use default JDM home from environment or current directory
		jdmHome := os.Getenv("JDM_HOME")
		if jdmHome == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				jdmHome = ".jdm"
			} else {
				jdmHome = fmt.Sprintf("%s/.jdm", home)
			}
		}
		Cfg = config.NewConfig(jdmHome)
		if err := Cfg.Load(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to load config: %v\n", err)
		}
	}
	return Cfg
}

// Execute runs the root command
func Execute() error {
	// Initialize config
	Cfg = getConfig()
	return rootCmd.Execute()
}

func init() {
	// Set version template for -v and --version flags
	rootCmd.SetVersionTemplate(`jdm v{{.Version}}\n`)
	rootCmd.Version = jdk.GetVersion()

	// jdm version command (as subcommand)
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print jdm version",
		Run: func(cmd *cobra.Command, args []string) {
			jdk.PrintVersion()
		},
	})

	// jdm help command - custom help that shows usage guide
	rootCmd.AddCommand(&cobra.Command{
		Use:   "help",
		Short: "Show jdm usage guide",
		Long: `JDM - JDK Version Manager

JDM is a tool for managing multiple JDK versions on Windows, similar to nvm.
It allows you to install, uninstall, and switch between different JDK versions.

EXAMPLES:
  jdm ls                       # List locally installed JDK versions
  jdm ls-remote               # List available JDK versions from remote
  jdm install 17              # Install JDK 17
  jdm use 17                  # Switch to JDK 17
  jdm default 17              # Set JDK 17 as default
  jdm alias set myjdk 17.0.2  # Create an alias for a version
  jdm config set mirror <url> # Set custom download mirror

GETTING STARTED:
  1. Run 'jdm install <version>' to install a JDK version
  2. Run 'jdm use <version>' to switch to the installed version
  3. Run 'jdm default <version>' to set a default version

For more information about each command, run 'jdm <command> --help'`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(cmd.Long)
		},
	})

	// Add log file flag
	rootCmd.PersistentFlags().StringVar(&logFile, "log", "", "log file path")

	// Set up PersistentPreRun to initialize logging
	cobra.OnInitialize(initLogging)
}

func initLogging() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			logrus.Warnf("Failed to open log file: %v", err)
		} else {
			logrus.SetOutput(f)
		}
	} else {
		// If no log file specified, try to initialize logging to default location
		// This uses the config if available, otherwise uses default JDM_HOME
		if Cfg != nil {
			if err := log.Init(Cfg.JDMHome); err != nil {
				logrus.Warnf("Failed to initialize log: %v", err)
			}
		}
	}
}

// Ensure basic commands work
func init() {
	// Add basic commands here
}
