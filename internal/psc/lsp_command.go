// ABOUTME: LSP command implementation for PSC
// ABOUTME: Provides language server functionality via command line interface

package psc

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/lsp"
)

// LSP command error codes
const (
	ErrLSPServerStart = "PSC-901" // Failed to start LSP server
	ErrLSPConnection  = "PSC-902" // LSP connection error
	ErrLSPInvalidPort = "PSC-903" // Invalid port number
)

// lspCmd represents the lsp command
var lspCmd = &cobra.Command{
	Use:   "lsp",
	Short: "Start PSC Language Server",
	Long: `Start the PSC Language Server for editor integration.

The language server provides type checking, hover information, and auto-completion
for editors and IDEs that support the Language Server Protocol (LSP).

Examples:
  # Start server using stdin/stdout (recommended for editors)
  psc lsp --stdio

  # Start server on TCP port for debugging
  psc lsp --tcp --port 9999

  # Start with verbose logging for debugging
  psc lsp --stdio --verbose`,
	RunE: runLSPCommand,
}

// LSP command options
type lspOptions struct {
	stdio   bool
	tcp     bool
	port    int
	verbose bool
	logFile string
}

var lspOpts = &lspOptions{}

func init() {
	// Add flags
	lspCmd.Flags().BoolVar(&lspOpts.stdio, "stdio", false, "Use stdin/stdout for communication (default for editors)")
	lspCmd.Flags().BoolVar(&lspOpts.tcp, "tcp", false, "Use TCP for communication (useful for debugging)")
	lspCmd.Flags().IntVar(&lspOpts.port, "port", 9999, "Port number for TCP mode")
	lspCmd.Flags().BoolVar(&lspOpts.verbose, "verbose", false, "Enable verbose logging")
	lspCmd.Flags().StringVar(&lspOpts.logFile, "log-file", "", "Log file path (default: stderr)")

	// Mark stdio and tcp as mutually exclusive
	lspCmd.MarkFlagsMutuallyExclusive("stdio", "tcp")

	// Add to PSC root command (done in init() of command.go)
}

// runLSPCommand executes the LSP server command
func runLSPCommand(cmd *cobra.Command, args []string) error {
	// Setup logging
	logger := setupLSPLogging(lspOpts)

	// Default to stdio if neither stdio nor tcp is specified
	if !lspOpts.stdio && !lspOpts.tcp {
		lspOpts.stdio = true
	}

	if lspOpts.stdio {
		return runStdioServer(logger)
	} else if lspOpts.tcp {
		return runTCPServer(logger)
	}

	return errors.NewUserInputError(
		ErrLSPServerStart,
		"Must specify either --stdio or --tcp mode",
		"",
		nil,
	)
}

// runStdioServer starts the LSP server in stdio mode
func runStdioServer(logger *log.Logger) error {
	logger.Println("Starting PSC Language Server in stdio mode")

	if err := lsp.StartStdioServer(); err != nil {
		return errors.NewSystemError(
			ErrLSPServerStart,
			"Failed to start LSP server in stdio mode",
			err,
		)
	}

	return nil
}

// runTCPServer starts the LSP server in TCP mode
func runTCPServer(logger *log.Logger) error {
	// Validate port number
	if lspOpts.port < 1 || lspOpts.port > 65535 {
		return errors.NewUserInputError(
			ErrLSPInvalidPort,
			fmt.Sprintf("Invalid port number: %d (must be 1-65535)", lspOpts.port),
			"",
			nil,
		)
	}

	address := fmt.Sprintf("localhost:%d", lspOpts.port)
	logger.Printf("Starting PSC Language Server on %s", address)

	// Start TCP server
	if err := lsp.StartTCPServer(address); err != nil {
		return errors.NewSystemError(
			ErrLSPServerStart,
			fmt.Sprintf("Failed to start LSP server on %s", address),
			err,
		)
	}

	return nil
}

// setupLSPLogging configures logging for the LSP server
func setupLSPLogging(opts *lspOptions) *log.Logger {
	var logOutput *os.File

	if opts.logFile != "" {
		// Log to specified file
		file, err := os.OpenFile(opts.logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			// Fall back to stderr if file can't be opened
			fmt.Fprintf(os.Stderr, "Warning: Could not open log file %s: %v\n", opts.logFile, err)
			logOutput = os.Stderr
		} else {
			logOutput = file
		}
	} else {
		// Log to stderr by default
		logOutput = os.Stderr
	}

	logger := log.New(logOutput, "[PSC-LSP] ", log.LstdFlags)

	if !opts.verbose {
		// In non-verbose mode, disable logging to avoid interfering with LSP communication
		// when using stdio mode
		if opts.stdio {
			logger.SetOutput(os.Stderr) // Always log to stderr in stdio mode
		}
	}

	return logger
}

// testLSPConnection tests the LSP server connection
func testLSPConnection(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return errors.NewSystemError(
			ErrLSPConnection,
			fmt.Sprintf("Failed to connect to LSP server at %s", address),
			err,
		)
	}
	defer func() { _ = conn.Close() }()
	return nil
}

// Additional helper commands for LSP integration

// lspTestCmd allows testing the LSP server functionality
var lspTestCmd = &cobra.Command{
	Use:   "test [address]",
	Short: "Test LSP server connection",
	Long: `Test connection to a running PSC Language Server.

Examples:
  # Test connection to server on default port
  psc lsp test

  # Test connection to server on specific port
  psc lsp test localhost:8080`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		address := "localhost:9999"
		if len(args) > 0 {
			address = args[0]
		}

		fmt.Printf("Testing connection to LSP server at %s...\n", address)

		if err := testLSPConnection(address); err != nil {
			fmt.Printf("❌ Connection failed: %v\n", err)
			return err
		}

		fmt.Printf("✅ Successfully connected to LSP server\n")
		return nil
	},
}

// lspStatusCmd shows the status of LSP-related services
var lspStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show LSP server status",
	Long:  `Show status information about PSC Language Server capabilities and configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("PSC Language Server Status")
		fmt.Println("==========================")
		fmt.Println()

		// Show server capabilities
		fmt.Println("Server Capabilities:")
		fmt.Println("✓ Text Document Synchronization (Full)")
		fmt.Println("✓ Hover Provider")
		fmt.Println("✓ Completion Provider")
		fmt.Println("✓ Diagnostics Provider")
		fmt.Println()

		// Show supported features
		fmt.Println("Supported Features:")
		fmt.Println("✓ Type checking with flow-sensitive analysis")
		fmt.Println("✓ Real-time error reporting")
		fmt.Println("✓ Auto-completion for types, keywords, and symbols")
		fmt.Println("✓ Hover information for variables and functions")
		fmt.Println("✓ Multiple diagnostic output formats (LSP, JSON, SARIF)")
		fmt.Println()

		// Show trigger characters
		fmt.Println("Completion Trigger Characters:")
		fmt.Println("$ @ % : . ->")
		fmt.Println()

		// Show supported file types
		fmt.Println("Supported File Types:")
		fmt.Println(".pl .pm .t")
		fmt.Println()

		return nil
	},
}

func init() {
	// Add subcommands to lsp command
	lspCmd.AddCommand(lspTestCmd)
	lspCmd.AddCommand(lspStatusCmd)
}
