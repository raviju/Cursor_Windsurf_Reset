package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"Cursor_Windsurf_Reset/cleaner"
	"Cursor_Windsurf_Reset/config"
	"Cursor_Windsurf_Reset/gui"
)

func main() {
	// 设置中文环境
	os.Setenv("LANG", "zh_CN.UTF-8")
	os.Setenv("LANGUAGE", "zh_CN.UTF-8")
	os.Setenv("LC_ALL", "zh_CN.UTF-8")

	// Fyne GUI设置
	os.Setenv("FYNE_FONT", "")      // 使用系统字体
	os.Setenv("FYNE_SCALE", "1.1")  // 使用更紧凑的界面比例
	os.Setenv("FYNE_THEME", "dark") // 使用暗色主题

	// Parse command line flags
	var (
		configPath = flag.String("config", "", "Configuration file path")
		discover   = flag.Bool("discover", false, "Discover and report application data locations")
		clean      = flag.String("clean", "", "Clean specific application (cursor/windsurf)")
		cleanAll   = flag.Bool("clean-all", false, "Clean all found applications")
		noConfirm  = flag.Bool("no-confirm", false, "Skip confirmation prompts")
		dryRun     = flag.Bool("dry-run", false, "Preview actions without making changes")
		verbose    = flag.Bool("verbose", false, "Show detailed output")
		cli        = flag.Bool("cli", false, "Use command line interface instead of GUI")
		version    = flag.Bool("version", false, "Show version information")
		testSQLite = flag.String("test-sqlite", "", "Test SQLite database connection (provide database path)")
	)
	flag.Parse()

	// Show version
	if *version {
		fmt.Println("Cursor & Windsurf Data Cleaner v2.0.0 (Go)")
		fmt.Println("Built with Go and Fyne GUI framework")
		return
	}

	// Setup logger
	logLevel := slog.LevelInfo
	if *verbose {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Create cleaning engine
	engine := cleaner.NewEngine(cfg, logger, *dryRun, *verbose)

	// Test SQLite connection if requested
	if *testSQLite != "" {
		fmt.Printf("Testing SQLite connection to: %s\n", *testSQLite)
		err := engine.TestSQLiteConnection(*testSQLite)
		if err != nil {
			fmt.Printf("❌ SQLite test failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ SQLite test successful")
		return
	}

	// CLI mode
	if *cli || *discover || *clean != "" || *cleanAll {
		runCLI(engine, cfg, logger, discover, clean, cleanAll, noConfirm, dryRun)
		return
	}

	// GUI mode
	runGUI()
}

// runCLI runs the command line interface
func runCLI(engine *cleaner.Engine, cfg *config.Config, logger *slog.Logger,
	discover *bool, clean *string, cleanAll *bool, noConfirm *bool, dryRun *bool) {

	fmt.Println("🧹 Cursor & Windsurf Data Cleaner v2.0.0 (Go)")
	fmt.Println(strings.Repeat("=", 55))
	fmt.Println("⚠️  IMPORTANT: This tool will modify application data.")
	fmt.Println("   Always backup your important work before proceeding.")
	fmt.Println("   Use this tool responsibly and in accordance with application ToS.")
	fmt.Println()

	// Discovery mode
	if *discover {
		performDiscovery(engine, cfg)
		return
	}

	// Get available applications
	appDataPaths := engine.GetAppDataPaths()
	availableApps := make([]string, 0)
	for appName, appPath := range appDataPaths {
		if appPath != "" {
			availableApps = append(availableApps, appName)
		}
	}

	if len(availableApps) == 0 {
		fmt.Println("❌ No supported applications found.")
		os.Exit(1)
	}

	// Determine which applications to clean
	var appsToClean []string

	if *clean != "" {
		found := false
		for _, app := range availableApps {
			if app == *clean {
				appsToClean = []string{app}
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("❌ Application '%s' not found or not supported.\n", *clean)
			os.Exit(1)
		}
	} else if *cleanAll {
		appsToClean = availableApps
	} else {
		// Interactive mode
		performDiscovery(engine, cfg)
		fmt.Println("\nAvailable applications to clean:")
		for i, app := range availableApps {
			appConfig := cfg.Applications[app]
			displayName := appConfig.DisplayName
			fmt.Printf("  %d. %s\n", i+1, displayName)
		}
		fmt.Println("  0. Exit")

		fmt.Print("\nSelect application to clean (number): ")
		var choice int
		fmt.Scanf("%d", &choice)

		if choice == 0 {
			return
		}

		if choice > 0 && choice <= len(availableApps) {
			appsToClean = []string{availableApps[choice-1]}
		} else {
			fmt.Println("❌ Invalid choice.")
			os.Exit(1)
		}
	}

	// Confirmation
	if !*noConfirm {
		safetyOptions := cfg.SafetyOptions
		if safetyOptions.RequireConfirmation {
			fmt.Printf("\n⚠️  You are about to clean data for: %s\n", appsToClean[0])
			fmt.Println("This will:")
			fmt.Println("  • Reset machine/device IDs")
			fmt.Println("  • Clear account-specific database records")
			fmt.Println("  • Remove cached workspace data")
			fmt.Println("  • Create backups of all modified files")

			fmt.Print("\nAre you sure you want to proceed? (type 'yes' to confirm): ")
			var confirm string
			fmt.Scanf("%s", &confirm)
			if confirm != "yes" {
				fmt.Println("Operation cancelled.")
				return
			}
		}
	}

	// Perform cleaning
	overallSuccess := true
	for _, appName := range appsToClean {
		fmt.Printf("\n🧹 Starting cleanup for %s...\n", appName)

		// Check if app is running
		if engine.IsAppRunning(appName) {
			fmt.Printf("❌ %s is currently running. Please close it first.\n", appName)
			overallSuccess = false
			continue
		}

		// Perform cleanup
		err := engine.CleanApplication(context.Background(), appName)
		if err != nil {
			fmt.Printf("❌ Failed to clean %s: %v\n", appName, err)
			overallSuccess = false
		} else {
			fmt.Printf("✅ Successfully cleaned %s\n", appName)
		}
	}

	// Summary
	fmt.Println("\n===== Cleaning Summary =====")
	if overallSuccess {
		fmt.Printf("✅ Successfully cleaned data for: %s\n", appsToClean[0])
		fmt.Printf("📁 Backups saved to: %s\n", engine.GetBackupDirectory())
		fmt.Println("\nYou can now launch the applications and log in with different accounts.")
	} else {
		fmt.Println("⚠️  Cleanup completed with some errors. Check the log for details.")
		fmt.Printf("📁 Backups saved to: %s\n", engine.GetBackupDirectory())
	}
}

// performDiscovery performs application discovery and reports results
func performDiscovery(engine *cleaner.Engine, cfg *config.Config) {
	fmt.Println("=== Application Data Discovery ===")

	appDataPaths := engine.GetAppDataPaths()
	for appName, appPath := range appDataPaths {
		appConfig := cfg.Applications[appName]
		displayName := appConfig.DisplayName

		if appPath != "" {
			fmt.Printf("%s: Found at %s\n", displayName, appPath)

			// Check if app is running
			if engine.IsAppRunning(appName) {
				fmt.Printf("  %s is currently running\n", displayName)
			} else {
				fmt.Printf("  %s is not running\n", displayName)
			}

			// Report size
			size := engine.GetDirectorySize(appPath)
			fmt.Printf("  💾 Size: %s\n", engine.FormatSize(size))
		} else {
			fmt.Printf("%s: Not found\n", displayName)
		}
	}

	fmt.Printf("📁 Backup directory: %s\n", engine.GetBackupDirectory())
}

// runGUI runs the graphical user interface
func runGUI() {
	app := gui.NewApp()
	app.Run()
}
