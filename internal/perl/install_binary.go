// ABOUTME: Binary installation functionality for Perl versions
// ABOUTME: Handles downloading, extracting, and installing pre-compiled Perl binaries

package perl

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/platform"
	"tamarou.com/pvm/internal/xdg"
)

// Binary installation error codes
const (
	ErrBinaryInstallFailed    = "701" // Binary installation failed
	ErrBinaryExtractFailed    = "702" // Binary extraction failed
	ErrBinaryVerifyFailed     = "703" // Binary verification failed
	ErrBinaryPermissionFailed = "704" // Failed to set permissions
)

// BinaryInstallOptions contains options for binary installation
type BinaryInstallOptions struct {
	// Version to install
	Version string

	// Platform triple (e.g., "linux-amd64", "darwin-arm64")
	Platform string

	// Installation directory (if empty, uses default)
	InstallDir string

	// Progress callback function
	ProgressCallback ProgressCallback

	// Context for cancellation
	Context context.Context

	// Skip checksum validation
	SkipChecksum bool

	// Custom repository URL
	RepoURL string
}

// BinaryInstallResult contains information about the binary installation
type BinaryInstallResult struct {
	// Version that was installed
	Version string

	// Platform that was installed
	Platform string

	// Installation path
	InstallPath string

	// Size of the installed files in bytes
	Size int64

	// Time taken to install
	Duration time.Duration

	// Whether installation was from cache
	FromCache bool
}

// InstallFromBinary installs Perl from a pre-compiled binary
func InstallFromBinary(options *BinaryInstallOptions) (*BinaryInstallResult, error) {
	startTime := time.Now()

	// Use default options if not specified
	if options == nil {
		options = &BinaryInstallOptions{}
	}

	// Set default platform if not specified
	if options.Platform == "" {
		options.Platform = platform.GetPlatformTriple()
	}

	// Set default context if not specified
	if options.Context == nil {
		options.Context = context.Background()
	}

	// Validate version
	parsedVersion, err := ParseVersion(options.Version)
	if err != nil {
		return nil, errors.NewVersionError(
			ErrInvalidBinaryVersion,
			fmt.Sprintf("Invalid version format: %s", options.Version),
			err)
	}
	version := parsedVersion.String()

	// Determine installation directory
	installDir := options.InstallDir
	if installDir == "" {
		dirs, err := xdg.GetDirs()
		if err != nil {
			return nil, errors.NewSystemError("001",
				"Failed to determine XDG directories", err)
		}

		installDir = filepath.Join(dirs.DataDir, "versions", version)
	}

	// Check if version is already installed
	if _, err := os.Stat(installDir); err == nil {
		// Directory exists, check if it's a valid installation
		perlBinary := filepath.Join(installDir, "bin", "perl")
		if _, err := os.Stat(perlBinary); err == nil {
			return &BinaryInstallResult{
				Version:     version,
				Platform:    options.Platform,
				InstallPath: installDir,
				Duration:    0,
				FromCache:   true,
			}, nil
		}
	}

	// Download the binary
	downloadOptions := &BinaryDownloadOptions{
		Version:          version,
		Platform:         options.Platform,
		ProgressCallback: options.ProgressCallback,
		SkipChecksum:     options.SkipChecksum,
		Context:          options.Context,
		RepoURL:          options.RepoURL,
	}

	downloadResult, err := DownloadPerlBinary(downloadOptions)
	if err != nil {
		return nil, errors.NewVersionError(
			ErrBinaryInstallFailed,
			fmt.Sprintf("Failed to download binary for Perl %s", version),
			err)
	}

	// Create installation directory
	err = os.MkdirAll(installDir, 0755)
	if err != nil {
		return nil, errors.NewSystemError(ErrBinaryInstallFailed,
			"Failed to create installation directory", err).
			WithLocation(installDir)
	}

	// Extract the binary archive
	extractedSize, err := extractBinaryArchive(downloadResult.Path, installDir, options.Platform)
	if err != nil {
		// Clean up on failure
		_ = os.RemoveAll(installDir)
		return nil, err
	}

	// Verify installation
	err = verifyBinaryInstallation(installDir, version)
	if err != nil {
		// Clean up on failure
		_ = os.RemoveAll(installDir)
		return nil, errors.NewVersionError(
			ErrBinaryVerifyFailed,
			fmt.Sprintf("Binary installation verification failed for Perl %s", version),
			err).WithLocation(installDir)
	}

	// Set proper permissions for Unix systems
	if runtime.GOOS != "windows" {
		err = setBinaryPermissions(installDir)
		if err != nil {
			return nil, errors.NewSystemError(ErrBinaryPermissionFailed,
				"Failed to set proper permissions on installed binary", err).
				WithLocation(installDir)
		}
	}

	// Register the installation
	versionInfo := VersionInfo{
		Version:     version,
		InstallPath: installDir,
		InstallTime: time.Now(),
		Source:      "binary",
	}

	err = RegisterVersion(versionInfo)
	if err != nil {
		// Don't fail the installation if registration fails, just warn
		// The installation is still functional
		fmt.Fprintf(os.Stderr, "Warning: Failed to register version %s: %v\n", version, err)
	}

	return &BinaryInstallResult{
		Version:     version,
		Platform:    options.Platform,
		InstallPath: installDir,
		Size:        extractedSize,
		Duration:    time.Since(startTime),
		FromCache:   downloadResult.FromCache,
	}, nil
}

// extractBinaryArchive extracts a binary archive to the installation directory
func extractBinaryArchive(archivePath, installDir, platform string) (int64, error) {
	// Determine extraction method based on platform
	if strings.HasPrefix(platform, "windows") {
		// Extract ZIP archive
		return extractZipArchive(archivePath, installDir)
	} else {
		// Extract tar.gz archive
		return extractTarGzArchive(archivePath, installDir)
	}
}

// extractTarGzArchive extracts a tar.gz archive
func extractTarGzArchive(archivePath, installDir string) (int64, error) {
	var totalSize int64

	// Open the archive file
	file, err := os.Open(archivePath)
	if err != nil {
		return 0, errors.NewSystemError(ErrBinaryExtractFailed,
			"Failed to open binary archive", err).
			WithLocation(archivePath)
	}
	defer file.Close()

	// Create gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return 0, errors.NewSystemError(ErrBinaryExtractFailed,
			"Failed to create gzip reader", err).
			WithLocation(archivePath)
	}
	defer gzReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzReader)

	// Extract files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, errors.NewSystemError(ErrBinaryExtractFailed,
				"Failed to read tar header", err).
				WithLocation(archivePath)
		}

		// Construct target path
		targetPath := filepath.Join(installDir, header.Name)

		// Security check: ensure the file is within the installation directory
		if !strings.HasPrefix(targetPath, filepath.Clean(installDir)+string(os.PathSeparator)) {
			return 0, errors.NewSystemError(ErrBinaryExtractFailed,
				fmt.Sprintf("Archive contains file outside installation directory: %s", header.Name),
				nil).WithLocation(archivePath)
		}

		// Extract based on file type
		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			err = os.MkdirAll(targetPath, os.FileMode(header.Mode))
			if err != nil {
				return 0, errors.NewSystemError(ErrBinaryExtractFailed,
					"Failed to create directory", err).
					WithLocation(targetPath)
			}

		case tar.TypeReg:
			// Extract regular file
			err = os.MkdirAll(filepath.Dir(targetPath), 0755)
			if err != nil {
				return 0, errors.NewSystemError(ErrBinaryExtractFailed,
					"Failed to create parent directory", err).
					WithLocation(filepath.Dir(targetPath))
			}

			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return 0, errors.NewSystemError(ErrBinaryExtractFailed,
					"Failed to create output file", err).
					WithLocation(targetPath)
			}

			size, err := io.Copy(outFile, tarReader)
			outFile.Close()
			if err != nil {
				return 0, errors.NewSystemError(ErrBinaryExtractFailed,
					"Failed to write file contents", err).
					WithLocation(targetPath)
			}

			totalSize += size

		case tar.TypeSymlink:
			// Create symlink
			err = os.Symlink(header.Linkname, targetPath)
			if err != nil {
				return 0, errors.NewSystemError(ErrBinaryExtractFailed,
					"Failed to create symlink", err).
					WithLocation(targetPath)
			}

		case tar.TypeLink:
			// Create hard link
			linkTarget := filepath.Join(installDir, header.Linkname)
			err = os.Link(linkTarget, targetPath)
			if err != nil {
				return 0, errors.NewSystemError(ErrBinaryExtractFailed,
					"Failed to create hard link", err).
					WithLocation(targetPath)
			}

		default:
			// Skip other file types (char devices, block devices, etc.)
			continue
		}
	}

	return totalSize, nil
}

// extractZipArchive extracts a ZIP archive
func extractZipArchive(archivePath, installDir string) (int64, error) {
	var totalSize int64

	// Open the ZIP archive
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return 0, errors.NewSystemError(ErrBinaryExtractFailed,
			"Failed to open ZIP archive", err).
			WithLocation(archivePath)
	}
	defer reader.Close()

	// Extract files
	for _, file := range reader.File {
		// Construct target path
		targetPath := filepath.Join(installDir, file.Name)

		// Security check: ensure the file is within the installation directory
		if !strings.HasPrefix(targetPath, filepath.Clean(installDir)+string(os.PathSeparator)) {
			return 0, errors.NewSystemError(ErrBinaryExtractFailed,
				fmt.Sprintf("Archive contains file outside installation directory: %s", file.Name),
				nil).WithLocation(archivePath)
		}

		// Create parent directories
		err = os.MkdirAll(filepath.Dir(targetPath), 0755)
		if err != nil {
			return 0, errors.NewSystemError(ErrBinaryExtractFailed,
				"Failed to create parent directory", err).
				WithLocation(filepath.Dir(targetPath))
		}

		// Skip directories
		if strings.HasSuffix(file.Name, "/") {
			continue
		}

		// Open file in archive
		srcFile, err := file.Open()
		if err != nil {
			return 0, errors.NewSystemError(ErrBinaryExtractFailed,
				"Failed to open file in archive", err).
				WithLocation(file.Name)
		}

		// Create output file
		outFile, err := os.Create(targetPath)
		if err != nil {
			srcFile.Close()
			return 0, errors.NewSystemError(ErrBinaryExtractFailed,
				"Failed to create output file", err).
				WithLocation(targetPath)
		}

		// Copy file contents
		size, err := io.Copy(outFile, srcFile)
		srcFile.Close()
		outFile.Close()
		if err != nil {
			return 0, errors.NewSystemError(ErrBinaryExtractFailed,
				"Failed to copy file contents", err).
				WithLocation(targetPath)
		}

		totalSize += size

		// Set file permissions (Unix-style permissions in ZIP)
		if file.Mode() != 0 {
			err = os.Chmod(targetPath, file.Mode())
			if err != nil {
				// Don't fail on permission errors on Windows
				if runtime.GOOS != "windows" {
					return 0, errors.NewSystemError(ErrBinaryExtractFailed,
						"Failed to set file permissions", err).
						WithLocation(targetPath)
				}
			}
		}
	}

	return totalSize, nil
}

// verifyBinaryInstallation checks if the binary installation is valid
func verifyBinaryInstallation(installDir, expectedVersion string) error {
	// Check that the perl binary exists
	perlBinary := filepath.Join(installDir, "bin", "perl")
	if _, err := os.Stat(perlBinary); os.IsNotExist(err) {
		return fmt.Errorf("perl binary not found at expected location: %s", perlBinary)
	}

	// Check that the binary is executable
	if runtime.GOOS != "windows" {
		fileInfo, err := os.Stat(perlBinary)
		if err != nil {
			return fmt.Errorf("failed to check perl binary: %w", err)
		}

		if fileInfo.Mode()&0111 == 0 {
			return fmt.Errorf("perl binary is not executable: %s", perlBinary)
		}
	}

	// TODO: Add version verification by executing the binary
	// This would require careful handling of environment variables
	// and might be better suited for a separate validation step

	return nil
}

// setBinaryPermissions sets proper permissions on Unix systems
func setBinaryPermissions(installDir string) error {
	// Walk through the installation directory and set permissions
	err := filepath.Walk(installDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip if it's a symlink
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		// Set permissions based on file type
		if info.IsDir() {
			// Directories: 755 (rwxr-xr-x)
			err = os.Chmod(path, 0755)
		} else {
			// Check if file should be executable
			relPath, err := filepath.Rel(installDir, path)
			if err != nil {
				return err
			}

			// Files in bin/ should be executable
			if strings.HasPrefix(relPath, "bin/") {
				err = os.Chmod(path, 0755) // rwxr-xr-x
			} else {
				err = os.Chmod(path, 0644) // rw-r--r--
			}
		}

		if err != nil {
			return fmt.Errorf("failed to set permissions on %s: %w", path, err)
		}

		return nil
	})

	return err
}
