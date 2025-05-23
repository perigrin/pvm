// ABOUTME: MCP server command implementation for PVM
// ABOUTME: Provides the mcp-server subcommand to start the MCP server

package pvm

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/config"
	"tamarou.com/pvm/internal/mcp"
)

// newMCPCommand creates the mcp-server subcommand
func newMCPCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp-server",
		Short: "Start the Model Context Protocol server",
		Long: `Start the MCP server that provides LLMs with:
- Perl code analysis using PVM's type system
- Semantic code search via embeddings
- Intelligent code generation with collaborative sampling
- Rich context awareness and project-scoped operations

The server auto-discovers Perl projects in the current directory tree
and provides tools for code analysis, search, and generation.`,
		RunE: runMCPServer,
	}

	// Add flags
	cmd.Flags().String("host", "", "Host address to bind to (default from config)")
	cmd.Flags().Int("port", 0, "Port to bind to (default from config)")
	cmd.Flags().Bool("no-auto-discover", false, "Disable automatic project discovery")
	cmd.Flags().String("config", "", "Path to configuration file")

	return cmd
}

// runMCPServer implements the mcp-server command
func runMCPServer(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := loadMCPConfig(cmd)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override config with command line flags
	if err := applyMCPFlags(cmd, cfg); err != nil {
		return fmt.Errorf("failed to apply command line flags: %w", err)
	}

	// Validate configuration
	if errors := cfg.Validate(); len(errors) > 0 {
		cmd.PrintErrln("Configuration validation errors:")
		for _, err := range errors {
			cmd.PrintErrln("  -", err)
		}
		return fmt.Errorf("configuration validation failed")
	}

	// Create and start MCP server
	server, err := mcp.NewServer(cfg)
	if err != nil {
		return fmt.Errorf("failed to create MCP server: %w", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	serverErrChan := make(chan error, 1)
	go func() {
		cmd.Printf("Starting MCP server on %s:%d\n", cfg.MCP.Host, cfg.MCP.Port)
		cmd.Printf("Auto-discover projects: %v\n", cfg.MCP.AutoDiscoverProjects)

		// Show discovered projects
		projects := server.GetProjects()
		if len(projects) > 0 {
			cmd.Println("Discovered projects:")
			for path, proj := range projects {
				cmd.Printf("  - %s (%s)\n", path, proj.ProjectType)
				if proj.HasVersion {
					cmd.Printf("    Perl version: %s\n", proj.PerlVersion)
				}
			}
		} else {
			cmd.Println("No projects discovered in current directory tree")
		}

		cmd.Println("MCP server is ready to accept connections...")

		if err := server.Start(ctx); err != nil {
			serverErrChan <- err
		}
	}()

	// Wait for shutdown signal or server error
	select {
	case <-sigChan:
		cmd.Println("\nReceived shutdown signal, stopping MCP server...")

		// Create timeout context for graceful shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		// Cancel server context
		cancel()

		// Wait for server to stop
		if err := server.Stop(shutdownCtx); err != nil {
			cmd.Printf("Error during server shutdown: %v\n", err)
			return err
		}

		cmd.Println("MCP server stopped gracefully")
		return nil

	case err := <-serverErrChan:
		cancel()
		return fmt.Errorf("MCP server error: %w", err)
	}
}

// loadMCPConfig loads the PVM configuration for MCP
func loadMCPConfig(cmd *cobra.Command) (*config.Config, error) {
	// Get config file path from flag
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, err
	}

	// Load configuration using PVM's config system
	var cfg *config.Config
	if configPath != "" {
		// Load from specific file
		cfg, err = config.ParseFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load config from %s: %w", configPath, err)
		}
	} else {
		// Load using standard PVM config resolution
		cfg, err = config.LoadEffectiveConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load PVM configuration: %w", err)
		}
	}

	// Ensure MCP config exists
	if cfg.MCP == nil {
		// Use default MCP config
		defaultCfg := config.NewDefaultConfig()
		cfg.MCP = defaultCfg.MCP
	}

	return cfg, nil
}

// applyMCPFlags applies command line flags to override configuration
func applyMCPFlags(cmd *cobra.Command, cfg *config.Config) error {
	// Override host if specified
	if cmd.Flags().Changed("host") {
		host, err := cmd.Flags().GetString("host")
		if err != nil {
			return err
		}
		cfg.MCP.Host = host
	}

	// Override port if specified
	if cmd.Flags().Changed("port") {
		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			return err
		}
		cfg.MCP.Port = port
	}

	// Override auto-discover if specified
	if cmd.Flags().Changed("no-auto-discover") {
		noAutoDiscover, err := cmd.Flags().GetBool("no-auto-discover")
		if err != nil {
			return err
		}
		cfg.MCP.AutoDiscoverProjects = !noAutoDiscover
	}

	return nil
}
