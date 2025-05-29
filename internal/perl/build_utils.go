// ABOUTME: Utility functions for enhanced build system
// ABOUTME: Provides archive extraction, command running, and other build utilities

package perl

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/ulikunitz/xz"
	"tamarou.com/pvm/internal/log"
)

// ArchiveExtractor handles archive extraction with progress
type ArchiveExtractor struct {
	logger *log.Logger
}

// NewArchiveExtractor creates a new archive extractor
func NewArchiveExtractor(logger *log.Logger) *ArchiveExtractor {
	return &ArchiveExtractor{logger: logger}
}

// ExtractWithProgress extracts an archive with progress reporting
func (ae *ArchiveExtractor) ExtractWithProgress(
	archivePath string,
	destDir string,
	ctx context.Context,
	progressCb func(current, total int64),
) (string, error) {
	// Open archive
	file, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Check file is readable
	_, err = file.Stat()
	if err != nil {
		return "", err
	}

	// Create appropriate reader
	var reader io.Reader
	switch {
	case strings.HasSuffix(archivePath, ".tar.gz"), strings.HasSuffix(archivePath, ".tgz"):
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return "", err
		}
		defer gzReader.Close()
		reader = gzReader

	case strings.HasSuffix(archivePath, ".tar.xz"):
		xzReader, err := xz.NewReader(file)
		if err != nil {
			return "", err
		}
		reader = xzReader

	default:
		return "", fmt.Errorf("unsupported archive format: %s", filepath.Base(archivePath))
	}

	// Count files first for accurate progress
	totalFiles := int64(0)
	countReader := tar.NewReader(reader)
	for {
		_, err := countReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			ae.logger.Debugf("Error counting files, using estimate: %v", err)
			totalFiles = 1000 // Estimate
			break
		}
		totalFiles++
	}

	// Reset file for extraction
	file.Seek(0, 0)

	// Recreate readers
	switch {
	case strings.HasSuffix(archivePath, ".tar.gz"), strings.HasSuffix(archivePath, ".tgz"):
		gzReader, _ := gzip.NewReader(file)
		defer gzReader.Close()
		reader = gzReader
	case strings.HasSuffix(archivePath, ".tar.xz"):
		xzReader, _ := xz.NewReader(file)
		reader = xzReader
	}

	// Extract files
	tarReader := tar.NewReader(reader)
	extractedFiles := int64(0)
	rootDir := ""

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		target := filepath.Join(destDir, header.Name)

		// Security check
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)) {
			return "", fmt.Errorf("illegal file path in archive: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return "", err
			}

			// Track root directory
			if rootDir == "" && filepath.Dir(header.Name) == "." {
				rootDir = target
			}

		case tar.TypeReg:
			if err := ae.extractFile(target, tarReader, header); err != nil {
				return "", err
			}

		case tar.TypeSymlink:
			if err := os.Symlink(header.Linkname, target); err != nil {
				return "", err
			}
		}

		extractedFiles++
		if progressCb != nil {
			progressCb(extractedFiles, totalFiles)
		}
	}

	return rootDir, nil
}

// extractFile extracts a single file from the archive
func (ae *ArchiveExtractor) extractFile(target string, reader io.Reader, header *tar.Header) error {
	// Create directory
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}

	// Create file
	file, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy content
	_, err = io.Copy(file, reader)
	return err
}

// CommandRunner runs commands with enhanced error handling and output capture
type CommandRunner struct {
	logger *log.Logger
}

// NewCommandRunner creates a new command runner
func NewCommandRunner(logger *log.Logger) *CommandRunner {
	return &CommandRunner{logger: logger}
}

// Run executes a command and returns output
func (cr *CommandRunner) Run(
	dir string,
	command string,
	args []string,
	ctx context.Context,
) (string, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	cr.logger.Debugf("Running command: %s %v in %s", command, args, dir)

	err := cmd.Run()

	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\nSTDERR:\n" + stderr.String()
	}

	if err != nil {
		return output, fmt.Errorf("%w\nOutput: %s", err, output)
	}

	return output, nil
}

// RunWithProgress executes a command with line-by-line progress updates
func (cr *CommandRunner) RunWithProgress(
	dir string,
	command string,
	args []string,
	ctx context.Context,
	progressCb func(line string, isError bool),
) (string, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir

	// Create pipes
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return "", err
	}

	// Capture output
	var output bytes.Buffer
	var outputMu sync.Mutex

	// Read stdout
	stdoutDone := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := scanner.Text()

			outputMu.Lock()
			output.WriteString(line + "\n")
			outputMu.Unlock()

			if progressCb != nil {
				progressCb(line, false)
			}
		}
		stdoutDone <- scanner.Err()
	}()

	// Read stderr
	stderrDone := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := scanner.Text()

			outputMu.Lock()
			output.WriteString("STDERR: " + line + "\n")
			outputMu.Unlock()

			if progressCb != nil {
				progressCb(line, true)
			}
		}
		stderrDone <- scanner.Err()
	}()

	// Wait for command
	cmdErr := cmd.Wait()

	// Wait for output readers
	<-stdoutDone
	<-stderrDone

	outputMu.Lock()
	finalOutput := output.String()
	outputMu.Unlock()

	if cmdErr != nil {
		return finalOutput, fmt.Errorf("%w", cmdErr)
	}

	return finalOutput, nil
}

// ParallelExecutor executes tasks in parallel with error handling
type ParallelExecutor struct {
	maxWorkers int
	logger     *log.Logger
}

// NewParallelExecutor creates a new parallel executor
func NewParallelExecutor(maxWorkers int, logger *log.Logger) *ParallelExecutor {
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}
	return &ParallelExecutor{
		maxWorkers: maxWorkers,
		logger:     logger,
	}
}

// Task represents a parallel task
type Task struct {
	Name string
	Fn   func() error
}

// Execute runs tasks in parallel
func (pe *ParallelExecutor) Execute(ctx context.Context, tasks []Task) error {
	if len(tasks) == 0 {
		return nil
	}

	// Create channels
	taskCh := make(chan Task, len(tasks))
	errCh := make(chan error, len(tasks))

	// Create wait group
	var wg sync.WaitGroup

	// Start workers
	workers := pe.maxWorkers
	if workers > len(tasks) {
		workers = len(tasks)
	}

	var cancelled atomic.Bool

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for task := range taskCh {
				// Check if cancelled
				if cancelled.Load() {
					return
				}

				select {
				case <-ctx.Done():
					cancelled.Store(true)
					return
				default:
				}

				pe.logger.Debugf("Worker %d executing task: %s", workerID, task.Name)

				if err := task.Fn(); err != nil {
					pe.logger.Errorf("Task %s failed: %v", task.Name, err)
					errCh <- fmt.Errorf("task %s: %w", task.Name, err)
					cancelled.Store(true)
					return
				}
			}
		}(i)
	}

	// Queue tasks
	for _, task := range tasks {
		taskCh <- task
	}
	close(taskCh)

	// Wait for completion
	wg.Wait()
	close(errCh)

	// Check for errors
	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("parallel execution failed: %v", errs[0])
	}

	return nil
}

// FileHasher computes file hashes for verification
type FileHasher struct {
	mu sync.Mutex
}

// HashFile computes SHA256 hash of a file
func (fh *FileHasher) HashFile(path string) (string, error) {
	fh.mu.Lock()
	defer fh.mu.Unlock()

	return calculateFileChecksum(path)
}

// HashDirectory computes hash of directory contents
func (fh *FileHasher) HashDirectory(dir string) (string, error) {
	fh.mu.Lock()
	defer fh.mu.Unlock()

	var hashes []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			hash, err := calculateFileChecksum(path)
			if err != nil {
				return err
			}

			relPath, _ := filepath.Rel(dir, path)
			hashes = append(hashes, fmt.Sprintf("%s:%s", relPath, hash))
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	// Combine hashes
	h := sha256.New()
	for _, hash := range hashes {
		h.Write([]byte(hash))
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
